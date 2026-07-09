package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/dao/erp_dao"
	"github.com/statistics/lib/db"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
	"github.com/statistics/lib/tiktokopenapi"
)

const (
	shopTypeFull = 1
	shopTypeHalf = 2
)

type options struct {
	configDir      string
	shopType       int
	status         int
	version        string
	sellerType     string
	workers        int
	batchSize      int
	limit          int
	skipExisting   bool
	retries        int
	retryWait      time.Duration
	requestTimeout time.Duration
	dryRun         bool
	verbose        bool
}

type stats struct {
	queued  int64
	success int64
	failed  int64
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	opt := parseOptions()
	if err := validateOptions(opt); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting tiktok category attr sync workers=%d batch_size=%d limit=%d shop_type=%d status=%d version=%s seller_type=%s skip_existing=%t dry_run=%t", opt.workers, opt.batchSize, opt.limit, opt.shopType, opt.status, valueForLog(opt.version), valueForLog(opt.sellerType), opt.skipExisting, opt.dryRun)
	initConfig(opt.configDir)
	log.Printf("config loaded")
	if err := validateTiktokConfig(); err != nil {
		log.Fatal(err)
	}
	log.Printf("tiktok config checked")
	db.Init("erp", conf.GetMysqlSource("erp"))
	if err := pingDB(15 * time.Second); err != nil {
		log.Fatal(err)
	}
	log.Printf("database connected")

	ctx := context.Background()
	jobChan := make(chan *erp_model.Temp1TiktokCategoryPath, opt.workers*2)
	stat := &stats{}

	var wg sync.WaitGroup
	for i := 0; i < opt.workers; i++ {
		wg.Add(1)
		go worker(ctx, i+1, opt, jobChan, stat, &wg)
	}

	progressDone := make(chan struct{})
	go printProgress(stat, progressDone)

	err := enqueueJobs(jobChan, opt, stat)
	close(jobChan)
	wg.Wait()
	close(progressDone)

	log.Printf("finished queued=%d success=%d failed=%d", atomic.LoadInt64(&stat.queued), atomic.LoadInt64(&stat.success), atomic.LoadInt64(&stat.failed))
	if err != nil {
		log.Fatal(err)
	}
	if atomic.LoadInt64(&stat.failed) > 0 {
		os.Exit(1)
	}
}

func parseOptions() options {
	opt := options{}
	flag.StringVar(&opt.configDir, "config-dir", "", "directory containing server.yaml; defaults to ./statistics-server when present")
	flag.IntVar(&opt.shopType, "shop-type", 0, "0=all, 1=full, 2=half")
	flag.IntVar(&opt.status, "status", erp_dao.TiktokCategoryPathStatusPending, "path status to process: 0=pending, 2=failed retry")
	flag.StringVar(&opt.version, "version", "", "path version to process: empty=all, v1, v2, null")
	flag.StringVar(&opt.sellerType, "seller-type", "", "half shop seller type to process: LOCAL or CROSS_BORDER")
	flag.IntVar(&opt.workers, "workers", 8, "concurrent TikTok API workers")
	flag.IntVar(&opt.batchSize, "batch-size", 500, "path rows loaded per database batch")
	flag.IntVar(&opt.limit, "limit", 0, "maximum rows to queue; 0 means no limit")
	flag.BoolVar(&opt.skipExisting, "skip-existing", false, "skip rows already present in temp1_tiktok_category_attr")
	flag.IntVar(&opt.retries, "retries", 2, "retry count for a failed TikTok request")
	flag.DurationVar(&opt.retryWait, "retry-wait", 2*time.Second, "base wait before retrying a failed TikTok request")
	flag.DurationVar(&opt.requestTimeout, "request-timeout", 180*time.Second, "timeout for each TikTok request")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "call API but do not write temp1_tiktok_category_attr")
	flag.BoolVar(&opt.verbose, "verbose", false, "log each category request start and finish")
	flag.Parse()
	opt.version = strings.ToLower(strings.TrimSpace(opt.version))
	opt.sellerType = normalizeSellerType(opt.sellerType)
	return opt
}

func validateOptions(opt options) error {
	if opt.shopType != 0 && opt.shopType != shopTypeFull && opt.shopType != shopTypeHalf {
		return fmt.Errorf("unsupported shop-type: %d", opt.shopType)
	}
	if opt.status != erp_dao.TiktokCategoryPathStatusPending && opt.status != erp_dao.TiktokCategoryPathStatusFailed {
		return fmt.Errorf("unsupported status: %d", opt.status)
	}
	if opt.version != "" && opt.version != "v1" && opt.version != "v2" && opt.version != "null" {
		return fmt.Errorf("unsupported version: %s", opt.version)
	}
	if opt.sellerType != "" && opt.sellerType != tiktokopenapi.SellerTypeLocal && opt.sellerType != tiktokopenapi.SellerTypeCrossBorder {
		return fmt.Errorf("unsupported seller-type: %s", opt.sellerType)
	}
	if opt.sellerType != "" && opt.shopType != shopTypeHalf {
		return fmt.Errorf("seller-type is only supported with shop-type=2")
	}
	if opt.shopType == shopTypeHalf {
		if opt.version != "v1" && opt.version != "v2" {
			return fmt.Errorf("shop-type=2 requires -version=v1 or -version=v2")
		}
		if opt.sellerType == "" {
			return fmt.Errorf("shop-type=2 requires -seller-type=LOCAL or -seller-type=CROSS_BORDER")
		}
	}
	if opt.workers <= 0 {
		return fmt.Errorf("workers must be greater than 0")
	}
	if opt.batchSize <= 0 {
		return fmt.Errorf("batch-size must be greater than 0")
	}
	if opt.limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if opt.retries < 0 {
		return fmt.Errorf("retries cannot be negative")
	}
	if opt.requestTimeout <= 0 {
		return fmt.Errorf("request-timeout must be greater than 0")
	}
	return nil
}

func initConfig(configDir string) {
	if configDir == "" {
		if exists(filepath.Join("statistics-server", "server.yaml")) {
			configDir = "statistics-server"
		} else if exists("server.yaml") {
			configDir = "."
		}
	}
	if configDir != "" {
		if abs, err := filepath.Abs(configDir); err == nil {
			configDir = abs
		}
		_ = os.Setenv("statistics-server-PATH", configDir)
	}
	conf.Init("statistics-server")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateTiktokConfig() error {
	tc := conf.GetTiktok()
	if hasTiktokCredentials(tc.AppKey, tc.AppSecret, tc.AccessToken) {
		return nil
	}
	if len(tc.Shops) == 0 {
		return fmt.Errorf("missing tiktok.app_key in server.yaml")
	}

	for name, shop := range tc.Shops {
		appKey := firstNonEmpty(shop.AppKey, tc.AppKey)
		appSecret := firstNonEmpty(shop.AppSecret, tc.AppSecret)
		accessToken := firstNonEmpty(shop.AccessToken, tc.AccessToken)
		if !hasTiktokCredentials(appKey, appSecret, accessToken) {
			return fmt.Errorf("missing tiktok.shops.%s app_key/app_secret/access_token in server.yaml", name)
		}
	}
	return nil
}

func hasTiktokCredentials(appKey, appSecret, accessToken string) bool {
	return appKey != "" && appSecret != "" && accessToken != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func pingDB(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.GetFactory("erp").DB.PingContext(ctx)
}

func enqueueJobs(jobChan chan<- *erp_model.Temp1TiktokCategoryPath, opt options, stat *stats) error {
	session := db.GetFactory("erp").GetSession()
	afterID := uint64(0)
	remaining := opt.limit

	for {
		batchSize := opt.batchSize
		if remaining > 0 && remaining < batchSize {
			batchSize = remaining
		}

		log.Printf("querying path batch after_id=%d batch_size=%d", afterID, batchSize)
		paths, err := erp_dao.Temp1TiktokCategoryPathDao.ListForAttrSync(session, erp_dao.TiktokCategoryPathAttrSyncFilter{
			ShopType:     opt.shopType,
			Status:       opt.status,
			Version:      opt.version,
			SellerType:   opt.sellerType,
			AfterID:      afterID,
			Limit:        batchSize,
			SkipExisting: opt.skipExisting,
		})
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			log.Printf("no more path rows after_id=%d", afterID)
			return nil
		}
		log.Printf("loaded path batch size=%d first_id=%d last_id=%d", len(paths), paths[0].Id, paths[len(paths)-1].Id)

		for _, path := range paths {
			afterID = path.Id
			jobChan <- path
			atomic.AddInt64(&stat.queued, 1)
			if remaining > 0 {
				remaining--
				if remaining == 0 {
					log.Printf("queued limit reached, last_id=%d", afterID)
					return nil
				}
			}
		}

		log.Printf("queued batch size=%d last_id=%d total_queued=%d", len(paths), afterID, atomic.LoadInt64(&stat.queued))
	}
}

func worker(ctx context.Context, workerID int, opt options, jobChan <-chan *erp_model.Temp1TiktokCategoryPath, stat *stats, wg *sync.WaitGroup) {
	defer wg.Done()

	session := db.GetFactory("erp").GetSession()
	for path := range jobChan {
		startTime := time.Now()
		if opt.verbose {
			log.Printf("worker=%d category_id=%s shop_type=%s version=%s seller_type=%s start", workerID, path.CategoryId, nullInt64String(path.ShopType), nullString(path.Version), nullString(path.SellerType))
		}
		if err := syncOne(ctx, session, opt, path); err != nil {
			atomic.AddInt64(&stat.failed, 1)
			log.Printf("worker=%d category_id=%s shop_type=%s version=%s seller_type=%s failed: %v", workerID, path.CategoryId, nullInt64String(path.ShopType), nullString(path.Version), nullString(path.SellerType), err)
			if !opt.dryRun {
				if statusErr := erp_dao.Temp1TiktokCategoryPathDao.UpdateStatusByID(session, path.Id, erp_dao.TiktokCategoryPathStatusFailed); statusErr != nil {
					log.Printf("worker=%d category_id=%s mark failed status error: %v", workerID, path.CategoryId, statusErr)
				}
			}
			continue
		}
		atomic.AddInt64(&stat.success, 1)
		if opt.verbose {
			log.Printf("worker=%d category_id=%s done cost=%s", workerID, path.CategoryId, time.Since(startTime))
		}
	}
}

func syncOne(ctx context.Context, session *dbsession.DBSession, opt options, path *erp_model.Temp1TiktokCategoryPath) error {
	shopType, ok := pathShopType(path)
	if !ok {
		return fmt.Errorf("invalid shop_type")
	}

	raw, err := fetchAttributeWithRetry(ctx, opt, path, shopType)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	if !json.Valid(raw) {
		return fmt.Errorf("tiktok response is not valid json")
	}
	if opt.dryRun {
		return nil
	}

	_, err = erp_dao.Temp1TiktokCategoryAttrDao.Upsert(session, &erp_model.Temp1TiktokCategoryAttr{
		CategoryId: path.CategoryId,
		RawJson:    raw,
		ShopType:   path.ShopType,
		Version:    path.Version,
		SellerType: path.SellerType,
		DeletedAt:  sql.NullTime{},
	})
	if err != nil {
		return err
	}
	return erp_dao.Temp1TiktokCategoryPathDao.UpdateStatusByID(session, path.Id, erp_dao.TiktokCategoryPathStatusSuccess)
}

func fetchAttributeWithRetry(ctx context.Context, opt options, path *erp_model.Temp1TiktokCategoryPath, shopType int) (json.RawMessage, error) {
	var lastErr error
	for attempt := 0; attempt <= opt.retries; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, opt.requestTimeout)
		raw, err := fetchAttribute(requestCtx, opt, path, shopType)
		cancel()
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if attempt < opt.retries {
			time.Sleep(opt.retryWait * time.Duration(attempt+1))
		}
	}
	return nil, lastErr
}

func fetchAttribute(ctx context.Context, opt options, path *erp_model.Temp1TiktokCategoryPath, shopType int) (json.RawMessage, error) {
	shopOptions, err := shopOptionsForPath(opt, path, shopType)
	if err != nil {
		return nil, err
	}
	client := tiktokopenapi.NewClientFromConf(shopOptions)
	service := client.NewCategoryService()

	switch shopType {
	case shopTypeFull:
		return service.GetFullCategoryAttributeList(ctx, path.CategoryId)
	case shopTypeHalf:
		return service.GetHalfCategoryAttributeList(ctx, path.CategoryId)
	default:
		return nil, fmt.Errorf("unsupported shop_type: %d", shopType)
	}
}

func pathShopType(path *erp_model.Temp1TiktokCategoryPath) (int, bool) {
	if !path.ShopType.Valid {
		return 0, false
	}
	shopType := int(path.ShopType.Int64)
	return shopType, shopType == shopTypeFull || shopType == shopTypeHalf
}

func shopOptionsForPath(opt options, path *erp_model.Temp1TiktokCategoryPath, shopType int) (tiktokopenapi.ShopOptions, error) {
	switch shopType {
	case shopTypeFull:
		return tiktokopenapi.ShopOptions{
			ShopType:   shopTypeFull,
			SellerType: "ALL",
		}, nil
	case shopTypeHalf:
		sellerType := sellerTypeFromPath(opt, path)
		if sellerType == "" {
			return tiktokopenapi.ShopOptions{}, fmt.Errorf("missing seller_type for half shop category path id=%d category_id=%s", path.Id, path.CategoryId)
		}
		categoryVersion := categoryVersionFromPath(opt, path)
		if categoryVersion == "" {
			return tiktokopenapi.ShopOptions{}, fmt.Errorf("missing version for half shop category path id=%d category_id=%s", path.Id, path.CategoryId)
		}
		return tiktokopenapi.ShopOptions{
			ShopType:        shopTypeHalf,
			SellerType:      sellerType,
			CategoryVersion: categoryVersion,
		}, nil
	default:
		return tiktokopenapi.ShopOptions{}, fmt.Errorf("unsupported shop_type: %d", shopType)
	}
}

func sellerTypeFromPath(opt options, path *erp_model.Temp1TiktokCategoryPath) string {
	if path.SellerType.Valid && path.SellerType.String != "" {
		return normalizeSellerType(path.SellerType.String)
	}
	return opt.sellerType
}

func categoryVersionFromPath(opt options, path *erp_model.Temp1TiktokCategoryPath) string {
	if path.Version.Valid && path.Version.String != "" {
		return strings.ToLower(strings.TrimSpace(path.Version.String))
	}
	return opt.version
}

func normalizeSellerType(sellerType string) string {
	sellerType = strings.ToUpper(strings.TrimSpace(sellerType))
	sellerType = strings.ReplaceAll(sellerType, "-", "_")
	return sellerType
}

func printProgress(stat *stats, done <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("progress queued=%d success=%d failed=%d", atomic.LoadInt64(&stat.queued), atomic.LoadInt64(&stat.success), atomic.LoadInt64(&stat.failed))
		case <-done:
			return
		}
	}
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return "NULL"
	}
	return value.String
}

func nullInt64String(value sql.NullInt64) string {
	if !value.Valid {
		return "NULL"
	}
	return fmt.Sprintf("%d", value.Int64)
}

func valueForLog(value string) string {
	if value == "" {
		return "all"
	}
	return value
}
