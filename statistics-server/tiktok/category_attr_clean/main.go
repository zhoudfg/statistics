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
	"github.com/statistics/lib/db"
)

const (
	shopTypeFull = 1
	shopTypeHalf = 2
)

type options struct {
	configDir       string
	shopType        int
	version         string
	sellerType      string
	workers         int
	batchSize       int
	limit           int
	afterID         uint64
	includeCleaned  bool
	replaceExisting bool
	dryRun          bool
	verbose         bool
}

type stats struct {
	seen         int64
	cleaned      int64
	failed       int64
	exportedRows int64
}

type idAllocator struct {
	next int64
}

type categoryAttrRow struct {
	AttrID     uint64
	CategoryID string
	RawJSON    json.RawMessage
	IsClean    sql.NullInt64
	ShopType   sql.NullInt64
	Version    sql.NullString
	SellerType sql.NullString
	PathCn     string
	PathID     string
	PathEn     string
}

type exportRow struct {
	ID                  int64
	ShopType            sql.NullInt64
	Version             sql.NullString
	SellerType          sql.NullString
	PathCn              sql.NullString
	PathID              sql.NullString
	PathEn              sql.NullString
	Attr                sql.NullString
	AttrID              sql.NullString
	AttrValues          sql.NullString
	AttrType            sql.NullString
	IsMultipleSelection sql.NullInt64
	IsRequired          sql.NullInt64
}

type rawAttributesPayload struct {
	Attributes []json.RawMessage `json:"attributes"`
}

type rawAttributeValue struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	NameZh string `json:"name_zh"`
}

type halfRawAttribute struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Type                string              `json:"type"`
	Values              []rawAttributeValue `json:"values"`
	IsRequried          *bool               `json:"is_requried"`
	IsRequired          *bool               `json:"is_required"`
	IsMultipleSelection *bool               `json:"is_multiple_selection"`
}

type fullRawAttribute struct {
	ID                  string              `json:"id"`
	Type                string              `json:"type"`
	Values              []rawAttributeValue `json:"values"`
	NameEn              string              `json:"name_en"`
	NameZh              string              `json:"name_zh"`
	ValueType           string              `json:"value_type"`
	IsRequired          *bool               `json:"is_required"`
	IsMutipleSelection  *bool               `json:"is_mutiple_selection"`
	IsMultipleSelection *bool               `json:"is_multiple_selection"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	opt := parseOptions()
	if err := validateOptions(opt); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting tiktok category attr clean workers=%d batch_size=%d limit=%d after_id=%d shop_type=%d version=%s seller_type=%s include_cleaned=%t replace_existing=%t dry_run=%t",
		opt.workers, opt.batchSize, opt.limit, opt.afterID, opt.shopType, valueForLog(opt.version), valueForLog(opt.sellerType), opt.includeCleaned, opt.replaceExisting, opt.dryRun)
	initConfig(opt.configDir)
	db.Init("erp", conf.GetMysqlSource("erp"))
	database := db.GetFactory("erp").DB
	database.SetMaxOpenConns(opt.workers + 4)
	database.SetMaxIdleConns(opt.workers + 4)
	if err := pingDB(database, 15*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Printf("database connected")

	nextExportID := int64(0)
	if !opt.dryRun {
		var err error
		nextExportID, err = getNextExportID(database)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("next export id=%d", nextExportID)
	}
	allocator := &idAllocator{next: nextExportID}

	stat := stats{}
	afterID := opt.afterID
	remaining := opt.limit

	for {
		batchSize := opt.batchSize
		if remaining > 0 && remaining < batchSize {
			batchSize = remaining
		}

		log.Printf("querying attr batch after_id=%d batch_size=%d", afterID, batchSize)
		rows, batchLastID, err := listRows(database, opt, afterID, batchSize)
		if err != nil {
			log.Fatal(err)
		}
		if batchLastID == 0 {
			log.Printf("no more attr rows after_id=%d", afterID)
			break
		}
		if len(rows) == 0 {
			afterID = batchLastID
			log.Printf("no joined path rows in batch, skip to after_id=%d", afterID)
			continue
		}

		batchStat := cleanRows(database, opt, rows, allocator)
		stat.seen += batchStat.seen
		stat.cleaned += batchStat.cleaned
		stat.failed += batchStat.failed
		stat.exportedRows += batchStat.exportedRows
		afterID = rows[len(rows)-1].AttrID

		if remaining > 0 {
			remaining -= len(rows)
			if remaining <= 0 {
				log.Printf("limit reached, last_attr_id=%d", afterID)
			}
		}
		if batchLastID > afterID {
			afterID = batchLastID
		}

		log.Printf("progress seen=%d cleaned=%d failed=%d exported_rows=%d last_attr_id=%d",
			stat.seen, stat.cleaned, stat.failed, stat.exportedRows, afterID)
		if remaining == 0 && opt.limit > 0 {
			break
		}
	}

	log.Printf("finished seen=%d cleaned=%d failed=%d exported_rows=%d", stat.seen, stat.cleaned, stat.failed, stat.exportedRows)
	if stat.failed > 0 {
		os.Exit(1)
	}
}

func parseOptions() options {
	opt := options{}
	flag.StringVar(&opt.configDir, "config-dir", "", "directory containing server.yaml; defaults to ./statistics-server when present")
	flag.IntVar(&opt.shopType, "shop-type", 0, "0=all, 1=full, 2=half")
	flag.StringVar(&opt.version, "version", "", "version to process: empty=all, v1, v2, null")
	flag.StringVar(&opt.sellerType, "seller-type", "", "seller type to process: LOCAL, CROSS_BORDER, ALL, null")
	flag.IntVar(&opt.workers, "workers", 8, "concurrent database cleaning workers")
	flag.IntVar(&opt.batchSize, "batch-size", 500, "attr rows loaded per database batch")
	flag.IntVar(&opt.limit, "limit", 0, "maximum attr rows to clean; 0 means no limit")
	flag.Uint64Var(&opt.afterID, "after-id", 0, "resume from temp1_tiktok_category_attr.id greater than this value")
	flag.BoolVar(&opt.includeCleaned, "include-cleaned", false, "also process rows whose is_clean = 1")
	flag.BoolVar(&opt.replaceExisting, "replace-existing", false, "soft delete existing export rows for the same category path before inserting")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "parse and count rows without writing temp1_tiktok_category_export or updating is_clean")
	flag.BoolVar(&opt.verbose, "verbose", false, "log each cleaned attr row")
	flag.Parse()
	opt.version = strings.ToLower(strings.TrimSpace(opt.version))
	opt.sellerType = normalizeSellerType(opt.sellerType)
	return opt
}

func validateOptions(opt options) error {
	if opt.shopType != 0 && opt.shopType != shopTypeFull && opt.shopType != shopTypeHalf {
		return fmt.Errorf("unsupported shop-type: %d", opt.shopType)
	}
	if opt.version != "" && opt.version != "v1" && opt.version != "v2" && opt.version != "null" {
		return fmt.Errorf("unsupported version: %s", opt.version)
	}
	if opt.sellerType != "" && opt.sellerType != "LOCAL" && opt.sellerType != "CROSS_BORDER" && opt.sellerType != "ALL" && opt.sellerType != "null" {
		return fmt.Errorf("unsupported seller-type: %s", opt.sellerType)
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

func listRows(database *sql.DB, opt options, afterID uint64, limit int) ([]categoryAttrRow, uint64, error) {
	ids, err := listAttrIDs(database, opt, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return nil, 0, nil
	}

	rows, err := listRowsByIDs(database, ids)
	if err != nil {
		return nil, 0, err
	}
	return rows, ids[len(ids)-1], nil
}

func listAttrIDs(database *sql.DB, opt options, afterID uint64, limit int) ([]uint64, error) {
	whereSQL := []string{
		"a.deleted_at IS NULL",
		"a.category_id <> ''",
	}
	args := make([]interface{}, 0)
	if !opt.includeCleaned {
		whereSQL = append(whereSQL, "(a.is_clean IS NULL OR a.is_clean = 0)")
	}
	if opt.shopType > 0 {
		whereSQL = append(whereSQL, "a.shop_type = ?")
		args = append(args, opt.shopType)
	}
	if opt.version != "" {
		if opt.version == "null" {
			whereSQL = append(whereSQL, "a.version IS NULL")
		} else {
			whereSQL = append(whereSQL, "a.version = ?")
			args = append(args, opt.version)
		}
	}
	if opt.sellerType != "" {
		if opt.sellerType == "null" {
			whereSQL = append(whereSQL, "a.seller_type IS NULL")
		} else {
			whereSQL = append(whereSQL, "a.seller_type = ?")
			args = append(args, opt.sellerType)
		}
	}
	if afterID > 0 {
		whereSQL = append(whereSQL, "a.id > ?")
		args = append(args, afterID)
	}
	args = append(args, limit)

	query := `SELECT a.id
FROM temp1_tiktok_category_attr a FORCE INDEX (PRIMARY)
WHERE ` + strings.Join(whereSQL, " AND ") + `
ORDER BY a.id ASC
LIMIT ?`

	sqlRows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	ids := make([]uint64, 0, limit)
	for sqlRows.Next() {
		var id uint64
		if err := sqlRows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := sqlRows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func listRowsByIDs(database *sql.DB, ids []uint64) ([]categoryAttrRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	query := `SELECT a.id, a.category_id, a.raw_json, a.is_clean, a.shop_type, a.version, a.seller_type,
       p.category_path_cn, p.category_path_id, p.category_path_en
FROM temp1_tiktok_category_attr a
INNER JOIN temp1_tiktok_category_path p
ON p.category_id = a.category_id
AND p.shop_type <=> a.shop_type
AND p.version <=> a.version
AND p.seller_type <=> a.seller_type
WHERE a.id IN (` + placeholders(len(ids)) + `)
AND a.deleted_at IS NULL
AND p.deleted_at IS NULL`

	sqlRows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	rowByID := make(map[uint64]categoryAttrRow, len(ids))
	for sqlRows.Next() {
		row := categoryAttrRow{}
		var rawJSON []byte
		if err := sqlRows.Scan(
			&row.AttrID,
			&row.CategoryID,
			&rawJSON,
			&row.IsClean,
			&row.ShopType,
			&row.Version,
			&row.SellerType,
			&row.PathCn,
			&row.PathID,
			&row.PathEn,
		); err != nil {
			return nil, err
		}
		row.RawJSON = append(row.RawJSON, rawJSON...)
		rowByID[row.AttrID] = row
	}
	if err := sqlRows.Err(); err != nil {
		return nil, err
	}

	rows := make([]categoryAttrRow, 0, len(rowByID))
	for _, id := range ids {
		row, ok := rowByID[id]
		if ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func getNextExportID(database *sql.DB) (int64, error) {
	var nextID int64
	err := database.QueryRow(`SELECT COALESCE(MAX(id), 0) + 1 FROM temp1_tiktok_category_export`).Scan(&nextID)
	return nextID, err
}

func cleanRows(database *sql.DB, opt options, rows []categoryAttrRow, allocator *idAllocator) stats {
	stat := stats{}
	if len(rows) == 0 {
		return stat
	}

	workerCount := opt.workers
	if workerCount > len(rows) {
		workerCount = len(rows)
	}

	jobs := make(chan categoryAttrRow, len(rows))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for row := range jobs {
				atomic.AddInt64(&stat.seen, 1)
				inserted, err := cleanOne(database, opt, row, allocator)
				if err != nil {
					atomic.AddInt64(&stat.failed, 1)
					log.Printf("worker=%d attr_id=%d category_id=%s shop_type=%s version=%s seller_type=%s clean failed: %v",
						workerID, row.AttrID, row.CategoryID, nullInt64ForLog(row.ShopType), nullStringForLog(row.Version), nullStringForLog(row.SellerType), err)
					continue
				}
				atomic.AddInt64(&stat.cleaned, 1)
				atomic.AddInt64(&stat.exportedRows, int64(inserted))
				if opt.verbose {
					log.Printf("worker=%d attr_id=%d category_id=%s exported_rows=%d", workerID, row.AttrID, row.CategoryID, inserted)
				}
			}
		}(i + 1)
	}

	for _, row := range rows {
		jobs <- row
	}
	close(jobs)
	wg.Wait()

	return stats{
		seen:         atomic.LoadInt64(&stat.seen),
		cleaned:      atomic.LoadInt64(&stat.cleaned),
		failed:       atomic.LoadInt64(&stat.failed),
		exportedRows: atomic.LoadInt64(&stat.exportedRows),
	}
}

func (a *idAllocator) nextID() int64 {
	return atomic.AddInt64(&a.next, 1) - 1
}

func cleanOne(database *sql.DB, opt options, row categoryAttrRow, allocator *idAllocator) (int, error) {
	exportRows, err := buildExportRows(row)
	if err != nil {
		return 0, err
	}
	if opt.dryRun {
		return len(exportRows), nil
	}

	tx, err := database.Begin()
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if opt.replaceExisting {
		if err := softDeleteExportRows(tx, row); err != nil {
			return 0, err
		}
	}
	for i := range exportRows {
		exportRows[i].ID = allocator.nextID()
	}
	if err := insertExportRows(tx, exportRows); err != nil {
		return 0, err
	}
	if err := markAttrClean(tx, row.AttrID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	committed = true
	return len(exportRows), nil
}

func softDeleteExportRows(tx *sql.Tx, row categoryAttrRow) error {
	_, err := tx.Exec(
		`UPDATE temp1_tiktok_category_export
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE shop_type <=> ?
AND version <=> ?
AND seller_type <=> ?
AND path_cn <=> ?
AND path_id <=> ?
AND path_en <=> ?
AND deleted_at IS NULL`,
		row.ShopType,
		row.Version,
		row.SellerType,
		nonNullString(row.PathCn),
		nonNullString(row.PathID),
		nonNullString(row.PathEn),
	)
	return err
}

func insertExportRows(tx *sql.Tx, rows []exportRow) error {
	if len(rows) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(`INSERT INTO temp1_tiktok_category_export
(id, shop_type, version, seller_type, path_cn, path_id, path_en, attr, attr_id, attr_values, attr_type, is_multiple_selection, is_required, deleted_at)
VALUES `)
	args := make([]interface{}, 0, len(rows)*13)
	for i, row := range rows {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)")
		args = append(args,
			row.ID,
			row.ShopType,
			row.Version,
			row.SellerType,
			row.PathCn,
			row.PathID,
			row.PathEn,
			row.Attr,
			row.AttrID,
			row.AttrValues,
			row.AttrType,
			row.IsMultipleSelection,
			row.IsRequired,
		)
	}

	_, err := tx.Exec(builder.String(), args...)
	return err
}

func markAttrClean(tx *sql.Tx, attrID uint64) error {
	_, err := tx.Exec(
		`UPDATE temp1_tiktok_category_attr
SET is_clean = 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`,
		attrID,
	)
	return err
}

func buildExportRows(row categoryAttrRow) ([]exportRow, error) {
	if !row.ShopType.Valid {
		return nil, fmt.Errorf("missing shop_type")
	}
	shopType := int(row.ShopType.Int64)
	if shopType != shopTypeFull && shopType != shopTypeHalf {
		return nil, fmt.Errorf("unsupported shop_type: %d", shopType)
	}

	payload := rawAttributesPayload{}
	if len(strings.TrimSpace(string(row.RawJSON))) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(row.RawJSON, &payload); err != nil {
		return nil, err
	}

	exportRows := make([]exportRow, 0, len(payload.Attributes))
	for idx, rawAttr := range payload.Attributes {
		base := baseExportRow(row)
		switch shopType {
		case shopTypeHalf:
			exportRow, err := buildHalfExportRow(base, rawAttr)
			if err != nil {
				return nil, fmt.Errorf("half attribute index=%d: %w", idx, err)
			}
			exportRows = append(exportRows, exportRow)
		case shopTypeFull:
			exportRow, err := buildFullExportRow(base, rawAttr)
			if err != nil {
				return nil, fmt.Errorf("full attribute index=%d: %w", idx, err)
			}
			exportRows = append(exportRows, exportRow)
		}
	}
	return exportRows, nil
}

func baseExportRow(row categoryAttrRow) exportRow {
	return exportRow{
		ShopType:   row.ShopType,
		Version:    row.Version,
		SellerType: row.SellerType,
		PathCn:     nonNullString(row.PathCn),
		PathID:     nonNullString(row.PathID),
		PathEn:     nonNullString(row.PathEn),
	}
}

func buildHalfExportRow(base exportRow, rawAttr json.RawMessage) (exportRow, error) {
	attr := halfRawAttribute{}
	if err := json.Unmarshal(rawAttr, &attr); err != nil {
		return exportRow{}, err
	}
	base.Attr = nonNullString(strings.TrimSpace(attr.Name))
	base.AttrID = nonNullString(strings.TrimSpace(attr.ID))
	base.AttrType = nonNullString(strings.TrimSpace(attr.Type))
	base.AttrValues = nullableString(formatAttrValues(attr.Values, func(value rawAttributeValue) string {
		return strings.TrimSpace(value.Name)
	}))
	base.IsRequired = boolPtrToNullTinyint(firstBool(attr.IsRequried, attr.IsRequired))
	base.IsMultipleSelection = boolPtrToNullTinyint(attr.IsMultipleSelection)
	return base, nil
}

func buildFullExportRow(base exportRow, rawAttr json.RawMessage) (exportRow, error) {
	attr := fullRawAttribute{}
	if err := json.Unmarshal(rawAttr, &attr); err != nil {
		return exportRow{}, err
	}
	base.Attr = nonNullString(firstNonEmpty(attr.NameZh, attr.NameEn))
	base.AttrID = nonNullString(strings.TrimSpace(attr.ID))
	base.AttrType = nonNullString(firstNonEmpty(attr.ValueType, attr.Type))
	base.AttrValues = nullableString(formatAttrValues(attr.Values, func(value rawAttributeValue) string {
		return firstNonEmpty(value.NameZh, value.NameEn, value.Name)
	}))
	base.IsRequired = boolPtrToNullTinyint(attr.IsRequired)
	base.IsMultipleSelection = boolPtrToNullTinyint(firstBool(attr.IsMutipleSelection, attr.IsMultipleSelection))
	return base, nil
}

func formatAttrValues(values []rawAttributeValue, nameFn func(rawAttributeValue) string) string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		name := strings.TrimSpace(nameFn(value))
		if name == "" {
			continue
		}
		names = append(names, "\u201c"+name+"\u201d")
	}
	if len(names) == 0 {
		return ""
	}
	return "\u3010" + strings.Join(names, "\uff0c") + "\u3011"
}

func boolPtrToNullTinyint(value *bool) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	if *value {
		return sql.NullInt64{Int64: 1, Valid: true}
	}
	return sql.NullInt64{Int64: 0, Valid: true}
}

func firstBool(values ...*bool) *bool {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeSellerType(sellerType string) string {
	sellerType = strings.ToUpper(strings.TrimSpace(sellerType))
	sellerType = strings.ReplaceAll(sellerType, "-", "_")
	return sellerType
}

func nullStringForLog(value sql.NullString) string {
	if !value.Valid {
		return "NULL"
	}
	return value.String
}

func nullInt64ForLog(value sql.NullInt64) string {
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
