package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/db"
)

const (
	shopTypeFull = 1
	shopTypeHalf = 2
)

type options struct {
	configDir         string
	shopType          int
	version           string
	sellerType        string
	separator         string
	status            int
	dbBatchSize       int
	skipMissingParent bool
	dryRun            bool
	verbose           bool
}

type stats struct {
	loaded        int
	leaf          int
	built         int
	saved         int
	missingParent int
	cycle         int
}

type categoryNode struct {
	CategoryID string
	ParentID   string
	IsLeaf     int
	NameCn     string
	NameEn     string
	ShopType   int
	Version    string
	SellerType string
}

type groupKey struct {
	ShopType   int
	Version    string
	SellerType string
}

type pathRow struct {
	CategoryID     string
	CategoryPathCn string
	CategoryPathID string
	CategoryPathEn string
	ShopType       int
	Version        string
	SellerType     string
	Status         int
}

type buildIssue struct {
	missingParent bool
	cycle         bool
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	opt := parseOptions()
	if err := validateOptions(opt); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting tiktok category path clean shop_type=%d version=%s seller_type=%s separator=%q status=%d db_batch_size=%d skip_missing_parent=%t dry_run=%t",
		opt.shopType, valueForLog(opt.version), valueForLog(opt.sellerType), opt.separator, opt.status, opt.dbBatchSize, opt.skipMissingParent, opt.dryRun)
	initConfig(opt.configDir)
	db.Init("erp", conf.GetMysqlSource("erp"))
	database := db.GetFactory("erp").DB
	if err := pingDB(database, 15*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Printf("database connected")

	nodes, err := listCategories(database, opt)
	if err != nil {
		log.Fatal(err)
	}
	stat := stats{loaded: len(nodes)}
	rows, buildStat := buildPathRows(nodes, opt)
	stat.leaf = buildStat.leaf
	stat.built = buildStat.built
	stat.missingParent = buildStat.missingParent
	stat.cycle = buildStat.cycle

	if opt.dryRun {
		log.Printf("finished loaded=%d leaf=%d built=%d saved=%d missing_parent=%d cycle=%d",
			stat.loaded, stat.leaf, stat.built, stat.saved, stat.missingParent, stat.cycle)
		return
	}

	saved, err := upsertPathRows(database, rows, opt)
	if err != nil {
		log.Fatal(err)
	}
	stat.saved = saved
	log.Printf("finished loaded=%d leaf=%d built=%d saved=%d missing_parent=%d cycle=%d",
		stat.loaded, stat.leaf, stat.built, stat.saved, stat.missingParent, stat.cycle)
}

func parseOptions() options {
	opt := options{}
	flag.StringVar(&opt.configDir, "config-dir", "", "directory containing server.yaml; defaults to ./statistics-server when present")
	flag.IntVar(&opt.shopType, "shop-type", 0, "0=all, 1=full, 2=half")
	flag.StringVar(&opt.version, "version", "", "category version to clean: empty=all, v1, v2")
	flag.StringVar(&opt.sellerType, "seller-type", "", "seller type to clean: empty=all, LOCAL, CROSS_BORDER, ALL")
	flag.StringVar(&opt.separator, "separator", " > ", "path separator")
	flag.IntVar(&opt.status, "status", 0, "status written to temp1_tiktok_category_path: 0=pending, 1=success, 2=failed")
	flag.IntVar(&opt.dbBatchSize, "db-batch-size", 500, "rows per database upsert batch")
	flag.BoolVar(&opt.skipMissingParent, "skip-missing-parent", false, "skip leaf paths whose parent chain is incomplete")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "build leaf paths without writing temp1_tiktok_category_path")
	flag.BoolVar(&opt.verbose, "verbose", false, "log incomplete paths and cycles")
	flag.Parse()
	opt.version = strings.ToLower(strings.TrimSpace(opt.version))
	opt.sellerType = normalizeSellerType(opt.sellerType)
	return opt
}

func validateOptions(opt options) error {
	if opt.shopType != 0 && opt.shopType != shopTypeFull && opt.shopType != shopTypeHalf {
		return fmt.Errorf("unsupported shop-type: %d", opt.shopType)
	}
	if opt.version != "" && opt.version != "v1" && opt.version != "v2" {
		return fmt.Errorf("unsupported version: %s", opt.version)
	}
	if opt.sellerType != "" && opt.sellerType != "LOCAL" && opt.sellerType != "CROSS_BORDER" && opt.sellerType != "ALL" {
		return fmt.Errorf("unsupported seller-type: %s", opt.sellerType)
	}
	if opt.separator == "" {
		return fmt.Errorf("separator cannot be empty")
	}
	if opt.status < 0 || opt.status > 2 {
		return fmt.Errorf("unsupported status: %d", opt.status)
	}
	if opt.dbBatchSize <= 0 {
		return fmt.Errorf("db-batch-size must be greater than 0")
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

func listCategories(database *sql.DB, opt options) ([]categoryNode, error) {
	whereSQL := []string{"deleted_at IS NULL", "category_id <> ''"}
	args := make([]interface{}, 0)
	if opt.shopType > 0 {
		whereSQL = append(whereSQL, "shop_type = ?")
		args = append(args, opt.shopType)
	}
	if opt.version != "" {
		whereSQL = append(whereSQL, "version = ?")
		args = append(args, opt.version)
	}
	if opt.sellerType != "" {
		whereSQL = append(whereSQL, "seller_type = ?")
		args = append(args, opt.sellerType)
	}

	query := `SELECT category_id, parent_id, is_leaf, name_cn, name_en, shop_type, version, seller_type
FROM temp1_tiktok_category
WHERE ` + strings.Join(whereSQL, " AND ") + `
ORDER BY shop_type ASC, version ASC, seller_type ASC, id ASC`

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]categoryNode, 0)
	for rows.Next() {
		var node categoryNode
		if err := rows.Scan(
			&node.CategoryID,
			&node.ParentID,
			&node.IsLeaf,
			&node.NameCn,
			&node.NameEn,
			&node.ShopType,
			&node.Version,
			&node.SellerType,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func buildPathRows(nodes []categoryNode, opt options) ([]pathRow, stats) {
	stat := stats{}
	grouped := make(map[groupKey]map[string]categoryNode)
	for _, node := range nodes {
		key := node.groupKey()
		if grouped[key] == nil {
			grouped[key] = make(map[string]categoryNode)
		}
		grouped[key][node.CategoryID] = node
	}

	paths := make([]pathRow, 0)
	for _, node := range nodes {
		if node.IsLeaf != 1 {
			continue
		}
		stat.leaf++
		ancestors, issue := ancestorsForLeaf(node, grouped[node.groupKey()])
		if issue.cycle {
			stat.cycle++
			if opt.verbose {
				log.Printf("cycle category_id=%s shop_type=%d version=%s seller_type=%s", node.CategoryID, node.ShopType, node.Version, node.SellerType)
			}
			continue
		}
		if issue.missingParent {
			stat.missingParent++
			if opt.verbose {
				log.Printf("missing parent category_id=%s parent_id=%s shop_type=%d version=%s seller_type=%s", node.CategoryID, node.ParentID, node.ShopType, node.Version, node.SellerType)
			}
			if opt.skipMissingParent {
				continue
			}
		}
		paths = append(paths, buildPathRow(node, ancestors, opt))
		stat.built++
	}
	return paths, stat
}

func ancestorsForLeaf(leaf categoryNode, byID map[string]categoryNode) ([]categoryNode, buildIssue) {
	reversed := make([]categoryNode, 0)
	seen := make(map[string]struct{})
	current := leaf
	for {
		if _, ok := seen[current.CategoryID]; ok {
			return nil, buildIssue{cycle: true}
		}
		seen[current.CategoryID] = struct{}{}
		reversed = append(reversed, current)

		parentID := strings.TrimSpace(current.ParentID)
		if parentID == "" || parentID == "0" {
			break
		}
		parent, ok := byID[parentID]
		if !ok {
			return reverseNodes(reversed), buildIssue{missingParent: true}
		}
		current = parent
	}
	return reverseNodes(reversed), buildIssue{}
}

func buildPathRow(leaf categoryNode, ancestors []categoryNode, opt options) pathRow {
	categoryIDs := make([]string, 0, len(ancestors))
	namesCn := make([]string, 0, len(ancestors))
	namesEn := make([]string, 0, len(ancestors))
	for _, node := range ancestors {
		categoryIDs = append(categoryIDs, node.CategoryID)
		namesCn = append(namesCn, strings.TrimSpace(node.NameCn))
		namesEn = append(namesEn, strings.TrimSpace(node.NameEn))
	}
	return pathRow{
		CategoryID:     leaf.CategoryID,
		CategoryPathCn: strings.Join(namesCn, opt.separator),
		CategoryPathID: strings.Join(categoryIDs, opt.separator),
		CategoryPathEn: strings.Join(namesEn, opt.separator),
		ShopType:       leaf.ShopType,
		Version:        leaf.Version,
		SellerType:     leaf.SellerType,
		Status:         opt.status,
	}
}

func upsertPathRows(database *sql.DB, rows []pathRow, opt options) (int, error) {
	total := 0
	for start := 0; start < len(rows); start += opt.dbBatchSize {
		end := start + opt.dbBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := upsertPathBatch(database, rows[start:end]); err != nil {
			return total, err
		}
		total += end - start
	}
	return total, nil
}

func upsertPathBatch(database *sql.DB, rows []pathRow) error {
	if len(rows) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(`INSERT INTO temp1_tiktok_category_path
(category_id, category_path_cn, category_path_id, category_path_en, shop_type, version, seller_type, status, deleted_at)
VALUES `)
	args := make([]interface{}, 0, len(rows)*9)
	for i, row := range rows {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, NULL)")
		args = append(args,
			row.CategoryID,
			row.CategoryPathCn,
			row.CategoryPathID,
			row.CategoryPathEn,
			row.ShopType,
			row.Version,
			row.SellerType,
			row.Status,
		)
	}
	builder.WriteString(` ON DUPLICATE KEY UPDATE
category_path_cn = VALUES(category_path_cn),
category_path_id = VALUES(category_path_id),
category_path_en = VALUES(category_path_en),
status = VALUES(status),
deleted_at = VALUES(deleted_at),
updated_at = CURRENT_TIMESTAMP`)

	_, err := database.Exec(builder.String(), args...)
	return err
}

func (n categoryNode) groupKey() groupKey {
	return groupKey{
		ShopType:   n.ShopType,
		Version:    n.Version,
		SellerType: n.SellerType,
	}
}

func reverseNodes(nodes []categoryNode) []categoryNode {
	for left, right := 0, len(nodes)-1; left < right; left, right = left+1, right-1 {
		nodes[left], nodes[right] = nodes[right], nodes[left]
	}
	return nodes
}

func normalizeSellerType(sellerType string) string {
	sellerType = strings.ToUpper(strings.TrimSpace(sellerType))
	sellerType = strings.ReplaceAll(sellerType, "-", "_")
	return sellerType
}

func valueForLog(value string) string {
	if value == "" {
		return "all"
	}
	return value
}
