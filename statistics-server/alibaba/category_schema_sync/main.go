package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/statistics/lib/alibabaopenapi"
	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/dao/erp_dao"
	"github.com/statistics/lib/db"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

var scenes = [...]string{"cbu", "industry", "popular"}

type options struct {
	configDir      string
	shop           string
	scene          string
	workers        int
	qps            int
	batchSize      int
	limit          int
	retries        int
	retryWait      time.Duration
	requestTimeout time.Duration
	dryRun         bool
	verbose        bool
}

type counts struct {
	queued  int64
	fetched int64
	written int64
	skipped int64
	failed  int64
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	opt := parseOptions()
	if err := validateOptions(opt); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, opt)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	var opt options
	flag.StringVar(&opt.configDir, "config-dir", "", "directory containing server.yaml")
	flag.StringVar(&opt.shop, "shop", "", "shop name from alibaba.shops")
	flag.StringVar(&opt.scene, "scene", "", "cbu, industry, or popular (empty processes all scenes)")
	flag.IntVar(&opt.workers, "workers", 4, "concurrent 1688 API workers")
	flag.IntVar(&opt.qps, "qps", 3, "maximum 1688 requests per second, including retries")
	flag.IntVar(&opt.batchSize, "batch-size", 500, "rows loaded per database batch")
	flag.IntVar(&opt.limit, "limit", 0, "maximum rows to process; 0 means no limit")
	flag.IntVar(&opt.retries, "retries", 2, "number of retries for transient API errors")
	flag.DurationVar(&opt.retryWait, "retry-wait", 2*time.Second, "base wait before retrying")
	flag.DurationVar(&opt.requestTimeout, "request-timeout", 30*time.Second, "timeout for each API request")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "call the API and validate rules without writing to MySQL")
	flag.BoolVar(&opt.verbose, "verbose", false, "log the result of each category")
	flag.Parse()
	opt.scene = alibabaopenapi.NormalizeScene(strings.ToLower(strings.TrimSpace(opt.scene)))
	return opt
}

func validateOptions(opt options) error {
	switch opt.scene {
	case "", "cbu", "industry", "popular":
	default:
		return fmt.Errorf("unsupported scene %q", opt.scene)
	}
	if opt.workers <= 0 || opt.batchSize <= 0 || opt.qps <= 0 || opt.qps > int(time.Second) {
		return fmt.Errorf("workers, batch-size and qps must be positive, with qps at most %d", time.Second.Nanoseconds())
	}
	if opt.limit < 0 || opt.retries < 0 || opt.retryWait <= 0 || opt.requestTimeout <= 0 {
		return fmt.Errorf("limit and retries must be non-negative, and retry-wait and request-timeout must be positive")
	}
	return nil
}

func initConfig(configDir string) error {
	if configDir == "" && os.Getenv("statistics-server-PATH") == "" {
		for _, dir := range []string{"statistics-server", "."} {
			if _, err := os.Stat(filepath.Join(dir, "server.yaml")); err == nil {
				configDir = dir
				break
			}
		}
	}
	if configDir != "" {
		abs, err := filepath.Abs(configDir)
		if err != nil {
			return err
		}
		if err := os.Setenv("statistics-server-PATH", abs); err != nil {
			return err
		}
	}
	conf.Init("statistics-server")
	return nil
}

func run(parent context.Context, opt options) error {
	if err := initConfig(opt.configDir); err != nil {
		return err
	}
	shop, err := conf.GetAlibaba().ResolveShop(opt.shop)
	if err != nil {
		return err
	}
	client, err := alibabaopenapi.NewClient(alibabaopenapi.Config{
		BaseURL:     shop.BaseURL,
		AppKey:      shop.AppKey,
		AppSecret:   shop.AppSecret,
		AccessToken: shop.AccessToken,
		Timeout:     opt.requestTimeout,
	})
	if err != nil {
		return err
	}
	dsn := conf.GetMysqlSource("erp")
	if dsn == "" {
		return fmt.Errorf("mysql.erp is required")
	}
	db.Init("erp", dsn)
	factory := db.GetFactory("erp")
	defer factory.Close()
	pingCtx, stopPing := context.WithTimeout(parent, 15*time.Second)
	err = factory.PingContext(pingCtx)
	stopPing()
	if err != nil {
		return fmt.Errorf("connect to ERP MySQL: %w", err)
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	interval := time.Second / time.Duration(opt.qps)
	if time.Second%time.Duration(opt.qps) != 0 {
		interval++
	}
	rate := time.NewTicker(interval)
	defer rate.Stop()
	jobs := make(chan *erp_model.TempAlibabaCategoryPath, opt.workers)
	stats := make(map[string]*counts, len(scenes))
	for _, scene := range scenes {
		stats[scene] = &counts{}
	}
	log.Printf("starting 1688 schema sync scene=%q workers=%d qps=%d limit=%d dry_run=%t", opt.scene, opt.workers, opt.qps, opt.limit, opt.dryRun)

	var wg sync.WaitGroup
	for i := 0; i < opt.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, cancel, factory.GetSession(), client, opt, rate.C, jobs, stats)
		}()
	}
	session := factory.GetSession()
	queueErr := enqueueJobs(ctx, session, opt, jobs, stats)
	close(jobs)
	wg.Wait()

	var failed int64
	for _, scene := range scenes {
		if opt.scene != "" && scene != opt.scene {
			continue
		}
		s := stats[scene]
		log.Printf("scene=%s queued=%d fetched=%d written=%d skipped=%d failed=%d", scene, s.queued, s.fetched, s.written, s.skipped, s.failed)
		failed += s.failed
	}
	if err := parent.Err(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return fmt.Errorf("1688 shop authorization failed, sync stopped")
	}
	if queueErr != nil {
		return queueErr
	}

	countCtx, stopCount := context.WithTimeout(ctx, 15*time.Second)
	defer stopCount()
	remaining, err := erp_dao.TempAlibabaCategoryPathDao.CountMissingSchema(countCtx, session, opt.scene)
	if err != nil {
		return fmt.Errorf("count missing schema: %w", err)
	}
	remaining["cbu"] += remaining["processing"]
	for _, scene := range scenes {
		if opt.scene == "" || scene == opt.scene {
			log.Printf("scene=%s remaining=%d", scene, remaining[scene])
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d rows failed, rerun to retry missing schemas", failed)
	}
	return nil
}

func enqueueJobs(ctx context.Context, session *dbsession.DBSession, opt options, jobs chan<- *erp_model.TempAlibabaCategoryPath, stats map[string]*counts) error {
	var afterID int64
	queued := 0
	for opt.limit == 0 || queued < opt.limit {
		batchSize := opt.batchSize
		if opt.limit > 0 && opt.limit-queued < batchSize {
			batchSize = opt.limit - queued
		}
		paths, err := erp_dao.TempAlibabaCategoryPathDao.ListMissingSchema(ctx, session, afterID, batchSize, opt.scene)
		if err != nil {
			return fmt.Errorf("read category paths after id %d: %w", afterID, err)
		}
		if len(paths) == 0 {
			return nil
		}
		for _, path := range paths {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobs <- path:
				stats[alibabaopenapi.NormalizeScene(path.Scene)].queued++
				queued++
				afterID = path.Id
			}
		}
		log.Printf("queued=%d last_id=%d", queued, afterID)
	}
	return nil
}

func worker(ctx context.Context, cancel context.CancelFunc, session *dbsession.DBSession, client *alibabaopenapi.Client, opt options, ticks <-chan time.Time, jobs <-chan *erp_model.TempAlibabaCategoryPath, stats map[string]*counts) {
	for path := range jobs {
		if ctx.Err() != nil {
			return
		}
		scene := alibabaopenapi.NormalizeScene(path.Scene)
		s := stats[scene]
		rule, err := fetchRule(ctx, client, opt, ticks, path)
		if err == nil {
			atomic.AddInt64(&s.fetched, 1)
			result := "validated"
			if !opt.dryRun {
				written, writeErr := erp_dao.TempAlibabaCategoryPathDao.SaveSchemaIfMissing(ctx, session, path, rule)
				if writeErr != nil {
					err = fmt.Errorf("write schema: %w", writeErr)
				} else if written {
					atomic.AddInt64(&s.written, 1)
					result = "written"
				} else {
					atomic.AddInt64(&s.skipped, 1)
					result = "skipped"
				}
			}
			if err == nil {
				if opt.verbose {
					log.Printf("id=%d category_id=%d scene=%s result=%s", path.Id, path.CategoryId, scene, result)
				}
				continue
			}
		}
		authFailed := alibabaopenapi.IsUnauthorized(err)
		if ctx.Err() != nil && !authFailed {
			return
		}
		atomic.AddInt64(&s.failed, 1)
		log.Printf("id=%d category_id=%d scene=%s failed: %v", path.Id, path.CategoryId, scene, err)
		saveCtx := ctx
		if authFailed {
			cancel()
			// Record the authorization error after stopping the other workers.
			saveCtx = context.Background()
		}
		if !opt.dryRun {
			errorCtx, stopSave := context.WithTimeout(saveCtx, 5*time.Second)
			_, saveErr := erp_dao.TempAlibabaCategoryPathDao.SaveSchemaError(errorCtx, session, path, alibabaopenapi.ErrorForStorage(err))
			stopSave()
			if saveErr != nil {
				log.Printf("id=%d category_id=%d scene=%s save error_msg failed: %v", path.Id, path.CategoryId, scene, saveErr)
			}
		}
		if authFailed {
			return
		}
	}
}

func fetchRule(ctx context.Context, client *alibabaopenapi.Client, opt options, ticks <-chan time.Time, path *erp_model.TempAlibabaCategoryPath) (json.RawMessage, error) {
	wait := opt.retryWait
	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticks:
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		requestCtx, cancel := context.WithTimeout(ctx, opt.requestTimeout)
		rule, err := client.GetPublishingRule(requestCtx, path.CategoryType, path.CategoryId, path.Scene)
		cancel()
		if err == nil || attempt == opt.retries || !alibabaopenapi.IsRetryable(err) {
			return rule, err
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		if wait < 30*time.Second/2 {
			wait *= 2
		} else if wait < 30*time.Second {
			wait = 30 * time.Second
		}
	}
}
