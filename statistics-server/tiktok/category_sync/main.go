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
	"strconv"
	"strings"
	"time"

	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/db"
	"github.com/statistics/lib/tiktokopenapi"
)

const (
	shopTypeFull = 1
	shopTypeHalf = 2

	localeZhCN = "zh-CN"
	localeEnUS = "en-US"
)

type options struct {
	configDir      string
	shopType       int
	version        string
	sellerType     string
	locale         string
	pageSize       int
	maxPages       int
	dbBatchSize    int
	requestTimeout time.Duration
	dryRun         bool
	verbose        bool
}

type stats struct {
	fetched int
	saved   int
	pages   int
}

type categoryRow struct {
	CategoryID         string
	ParentID           string
	IsLeaf             int
	NameCn             string
	NameEn             string
	PermissionStatuses interface{}
	ShopType           int
	Status             sql.NullString
	SellerType         string
	Version            string
}

type halfCategoryResponse struct {
	Categories []halfCategory `json:"categories"`
}

type halfCategory struct {
	ID                 string          `json:"id"`
	ParentID           string          `json:"parent_id"`
	LocalName          string          `json:"local_name"`
	IsLeaf             bool            `json:"is_leaf"`
	PermissionStatuses json.RawMessage `json:"permission_statuses"`
}

type fullCategoryResponse struct {
	Categorys  []fullCategory `json:"categorys"`
	Categories []fullCategory `json:"categories"`
	PageOffset int            `json:"page_offset"`
	PageSize   int            `json:"page_size"`
	TotalCount int            `json:"total_count"`
}

type fullCategory struct {
	ID       string          `json:"id"`
	NameCn   string          `json:"name_cn"`
	NameEn   string          `json:"name_en"`
	ParentID string          `json:"parent_id"`
	IsLeaf   bool            `json:"is_leaf"`
	Status   json.RawMessage `json:"status"`
	Version  string          `json:"version"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	opt := parseOptions()
	if err := validateOptions(opt); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting tiktok category sync shop_type=%d version=%s seller_type=%s locale=%s page_size=%d max_pages=%d db_batch_size=%d dry_run=%t",
		opt.shopType, valueForLog(opt.version), valueForLog(opt.sellerType), opt.locale, opt.pageSize, opt.maxPages, opt.dbBatchSize, opt.dryRun)
	initConfig(opt.configDir)
	db.Init("erp", conf.GetMysqlSource("erp"))
	database := db.GetFactory("erp").DB
	if err := pingDB(database, 15*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Printf("database connected")

	ctx := context.Background()
	var stat stats
	var err error
	switch opt.shopType {
	case shopTypeHalf:
		stat, err = syncHalfCategories(ctx, database, opt)
	case shopTypeFull:
		stat, err = syncFullCategories(ctx, database, opt)
	default:
		err = fmt.Errorf("unsupported shop-type: %d", opt.shopType)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("finished fetched=%d saved=%d pages=%d", stat.fetched, stat.saved, stat.pages)
}

func parseOptions() options {
	opt := options{}
	flag.StringVar(&opt.configDir, "config-dir", "", "directory containing server.yaml; defaults to ./statistics-server when present")
	flag.IntVar(&opt.shopType, "shop-type", 0, "required: 1=full, 2=half")
	flag.StringVar(&opt.version, "version", "", "category version: v1 or v2")
	flag.StringVar(&opt.sellerType, "seller-type", "", "half shop seller type: LOCAL or CROSS_BORDER")
	flag.StringVar(&opt.locale, "locale", localeZhCN, "half category locale: zh-CN or en-US")
	flag.IntVar(&opt.pageSize, "page-size", 50, "full category page size; maximum 50")
	flag.IntVar(&opt.maxPages, "max-pages", 0, "maximum full category pages to fetch; 0 means no limit")
	flag.IntVar(&opt.dbBatchSize, "db-batch-size", 500, "rows per database upsert batch")
	flag.DurationVar(&opt.requestTimeout, "request-timeout", 180*time.Second, "timeout for each TikTok request")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "call API and parse rows without writing temp1_tiktok_category")
	flag.BoolVar(&opt.verbose, "verbose", false, "log page and row details")
	flag.Parse()
	opt.version = strings.ToLower(strings.TrimSpace(opt.version))
	opt.sellerType = normalizeSellerType(opt.sellerType)
	opt.locale = normalizeLocale(opt.locale)
	return opt
}

func validateOptions(opt options) error {
	if opt.shopType != shopTypeFull && opt.shopType != shopTypeHalf {
		return fmt.Errorf("shop-type is required: 1=full, 2=half")
	}
	if opt.version != "" && opt.version != tiktokopenapi.CategoryVersionV1 && opt.version != tiktokopenapi.CategoryVersionV2 {
		return fmt.Errorf("unsupported version: %s", opt.version)
	}
	if opt.shopType == shopTypeHalf {
		if opt.version == "" {
			return fmt.Errorf("shop-type=2 requires -version=v1 or -version=v2")
		}
		if opt.sellerType != tiktokopenapi.SellerTypeLocal && opt.sellerType != tiktokopenapi.SellerTypeCrossBorder {
			return fmt.Errorf("shop-type=2 requires -seller-type=LOCAL or -seller-type=CROSS_BORDER")
		}
		if opt.locale != localeZhCN && opt.locale != localeEnUS {
			return fmt.Errorf("unsupported locale for half category sync: %s", opt.locale)
		}
	}
	if opt.shopType == shopTypeFull && opt.sellerType != "" && opt.sellerType != "ALL" {
		return fmt.Errorf("shop-type=1 only supports empty seller-type or ALL")
	}
	if opt.pageSize <= 0 || opt.pageSize > 50 {
		return fmt.Errorf("page-size must be between 1 and 50")
	}
	if opt.maxPages < 0 {
		return fmt.Errorf("max-pages cannot be negative")
	}
	if opt.dbBatchSize <= 0 {
		return fmt.Errorf("db-batch-size must be greater than 0")
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

func pingDB(database *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return database.PingContext(ctx)
}

func syncHalfCategories(ctx context.Context, database *sql.DB, opt options) (stats, error) {
	requestCtx, cancel := context.WithTimeout(ctx, opt.requestTimeout)
	defer cancel()

	service := categoryServiceForOptions(opt)
	raw, err := service.GetHalfCategoryListWithLocale(requestCtx, opt.locale)
	if err != nil {
		return stats{}, err
	}

	payload := halfCategoryResponse{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return stats{}, err
	}

	rows := make([]categoryRow, 0, len(payload.Categories))
	for _, category := range payload.Categories {
		row := buildHalfCategoryRow(category, opt)
		if row.CategoryID == "" {
			continue
		}
		rows = append(rows, row)
	}
	if opt.dryRun {
		return stats{fetched: len(rows), saved: len(rows), pages: 1}, nil
	}
	saved, err := upsertCategoryRows(database, rows, opt)
	if err != nil {
		return stats{}, err
	}
	return stats{fetched: len(rows), saved: saved, pages: 1}, nil
}

func syncFullCategories(ctx context.Context, database *sql.DB, opt options) (stats, error) {
	service := categoryServiceForOptions(opt)
	stat := stats{}
	for pageOffset := 1; ; pageOffset++ {
		requestCtx, cancel := context.WithTimeout(ctx, opt.requestTimeout)
		raw, err := service.GetFullCategoryList(requestCtx, fullCategoryRequestParams(opt, pageOffset))
		cancel()
		if err != nil {
			return stat, err
		}

		payload := fullCategoryResponse{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return stat, err
		}
		stat.pages++
		pageCategories := payload.fullCategories()
		if opt.verbose {
			log.Printf("loaded full category page=%d page_size=%d total_count=%d rows=%d", pageOffset, payload.PageSize, payload.TotalCount, len(pageCategories))
		}
		if len(pageCategories) == 0 {
			break
		}

		rows := make([]categoryRow, 0, len(pageCategories))
		for _, category := range pageCategories {
			row := buildFullCategoryRow(category, opt)
			if row.CategoryID == "" {
				continue
			}
			if opt.version != "" && row.Version != opt.version {
				continue
			}
			rows = append(rows, row)
		}
		stat.fetched += len(rows)
		if opt.dryRun {
			stat.saved += len(rows)
		} else {
			saved, err := upsertCategoryRows(database, rows, opt)
			if err != nil {
				return stat, err
			}
			stat.saved += saved
		}

		if payload.TotalCount > 0 && pageOffset*opt.pageSize >= payload.TotalCount {
			break
		}
		if opt.maxPages > 0 && stat.pages >= opt.maxPages {
			break
		}
	}
	return stat, nil
}

func categoryServiceForOptions(opt options) *tiktokopenapi.CategoryService {
	shop := tiktokopenapi.ShopOptions{
		ShopType:        opt.shopType,
		CategoryVersion: opt.version,
	}
	if opt.shopType == shopTypeFull {
		shop.SellerType = "ALL"
	} else {
		shop.SellerType = opt.sellerType
	}
	return tiktokopenapi.NewClientFromConf(shop).NewCategoryService()
}

func fullCategoryRequestParams(opt options, pageOffset int) map[string]interface{} {
	return map[string]interface{}{
		"page_offset": pageOffset,
		"page_size":   opt.pageSize,
	}
}

func (r fullCategoryResponse) fullCategories() []fullCategory {
	if len(r.Categorys) > 0 {
		return r.Categorys
	}
	return r.Categories
}

func buildHalfCategoryRow(category halfCategory, opt options) categoryRow {
	row := categoryRow{
		CategoryID:         strings.TrimSpace(category.ID),
		ParentID:           strings.TrimSpace(category.ParentID),
		IsLeaf:             boolToTinyint(category.IsLeaf),
		PermissionStatuses: rawJSONOrNil(category.PermissionStatuses),
		ShopType:           shopTypeHalf,
		Status:             nonNullString(""),
		SellerType:         opt.sellerType,
		Version:            opt.version,
	}
	switch opt.locale {
	case localeZhCN:
		row.NameCn = strings.TrimSpace(category.LocalName)
	case localeEnUS:
		row.NameEn = strings.TrimSpace(category.LocalName)
	}
	return row
}

func buildFullCategoryRow(category fullCategory, opt options) categoryRow {
	version := strings.TrimSpace(category.Version)
	if version == "" {
		version = opt.version
	}
	return categoryRow{
		CategoryID:         strings.TrimSpace(category.ID),
		ParentID:           strings.TrimSpace(category.ParentID),
		IsLeaf:             boolToTinyint(category.IsLeaf),
		NameCn:             strings.TrimSpace(category.NameCn),
		NameEn:             strings.TrimSpace(category.NameEn),
		PermissionStatuses: nil,
		ShopType:           shopTypeFull,
		Status:             nullableString(rawJSONScalarToString(category.Status)),
		SellerType:         "ALL",
		Version:            version,
	}
}

func upsertCategoryRows(database *sql.DB, rows []categoryRow, opt options) (int, error) {
	total := 0
	for start := 0; start < len(rows); start += opt.dbBatchSize {
		end := start + opt.dbBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := upsertCategoryBatch(database, rows[start:end], opt); err != nil {
			return total, err
		}
		total += end - start
	}
	return total, nil
}

func upsertCategoryBatch(database *sql.DB, rows []categoryRow, opt options) error {
	if len(rows) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(`INSERT INTO temp1_tiktok_category
(category_id, parent_id, is_leaf, name_cn, name_en, permission_statuses, shop_type, status, seller_type, version, deleted_at)
VALUES `)
	args := make([]interface{}, 0, len(rows)*11)
	for i, row := range rows {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)")
		args = append(args,
			row.CategoryID,
			row.ParentID,
			row.IsLeaf,
			row.NameCn,
			row.NameEn,
			row.PermissionStatuses,
			row.ShopType,
			row.Status,
			row.SellerType,
			row.Version,
		)
	}
	builder.WriteString(" ON DUPLICATE KEY UPDATE parent_id = VALUES(parent_id), is_leaf = VALUES(is_leaf), permission_statuses = VALUES(permission_statuses), status = VALUES(status), deleted_at = VALUES(deleted_at), updated_at = CURRENT_TIMESTAMP")
	switch {
	case opt.shopType == shopTypeFull:
		builder.WriteString(", name_cn = VALUES(name_cn), name_en = VALUES(name_en)")
	case opt.shopType == shopTypeHalf && opt.locale == localeZhCN:
		builder.WriteString(", name_cn = VALUES(name_cn)")
	case opt.shopType == shopTypeHalf && opt.locale == localeEnUS:
		builder.WriteString(", name_en = VALUES(name_en)")
	}

	_, err := database.Exec(builder.String(), args...)
	return err
}

func rawJSONOrNil(raw json.RawMessage) interface{} {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return []byte(raw)
}

func rawJSONScalarToString(raw json.RawMessage) string {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return strings.TrimSpace(str)
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err == nil {
		return number.String()
	}
	var boolean bool
	if err := json.Unmarshal(raw, &boolean); err == nil {
		return strconv.FormatBool(boolean)
	}
	return strings.Trim(value, `"`)
}

func boolToTinyint(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nonNullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: true}
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func normalizeSellerType(sellerType string) string {
	sellerType = strings.ToUpper(strings.TrimSpace(sellerType))
	sellerType = strings.ReplaceAll(sellerType, "-", "_")
	return sellerType
}

func normalizeLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "", "zh-cn", "zh_cn":
		return localeZhCN
	case "en-us", "en_us":
		return localeEnUS
	default:
		return strings.TrimSpace(locale)
	}
}

func valueForLog(value string) string {
	if value == "" {
		return "all"
	}
	return value
}
