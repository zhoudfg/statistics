package erp_dao

import (
	"database/sql"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

type temp1TiktokCategoryPathDao struct{}

var Temp1TiktokCategoryPathDao = new(temp1TiktokCategoryPathDao)

const (
	TiktokCategoryPathStatusPending = 0
	TiktokCategoryPathStatusSuccess = 1
	TiktokCategoryPathStatusFailed  = 2
)

type TiktokCategoryPathAttrSyncFilter struct {
	ShopType     int
	Status       int
	Version      string
	SellerType   string
	AfterID      uint64
	Limit        int
	SkipExisting bool
}

func (*temp1TiktokCategoryPathDao) Insert(session *dbsession.DBSession, path *erp_model.Temp1TiktokCategoryPath) (int64, error) {
	result, err := session.Exec(
		`INSERT INTO temp1_tiktok_category_path
(category_id, category_path_cn, category_path_id, category_path_en, shop_type, version, seller_type, status, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		path.CategoryId,
		path.CategoryPathCn,
		path.CategoryPathId,
		path.CategoryPathEn,
		path.ShopType,
		path.Version,
		path.SellerType,
		path.Status,
		path.DeletedAt,
	)
	if err != nil {
		log.Error(err)
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return id, nil
}

func (*temp1TiktokCategoryPathDao) Upsert(session *dbsession.DBSession, path *erp_model.Temp1TiktokCategoryPath) (int64, error) {
	result, err := session.Exec(
		`INSERT INTO temp1_tiktok_category_path
(category_id, category_path_cn, category_path_id, category_path_en, shop_type, version, seller_type, status, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE category_path_cn = VALUES(category_path_cn), category_path_id = VALUES(category_path_id), category_path_en = VALUES(category_path_en), status = VALUES(status), deleted_at = VALUES(deleted_at), updated_at = CURRENT_TIMESTAMP`,
		path.CategoryId,
		path.CategoryPathCn,
		path.CategoryPathId,
		path.CategoryPathEn,
		path.ShopType,
		path.Version,
		path.SellerType,
		path.Status,
		path.DeletedAt,
	)
	if err != nil {
		log.Error(err)
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return id, nil
}

func (*temp1TiktokCategoryPathDao) GetByUniqueKey(session *dbsession.DBSession, categoryId string, shopType sql.NullInt64, version sql.NullString, sellerType sql.NullString) (*erp_model.Temp1TiktokCategoryPath, error) {
	row := session.QueryRow(
		`SELECT id, category_id, category_path_cn, category_path_id, category_path_en, shop_type, version, seller_type, status, created_at, updated_at, deleted_at
FROM temp1_tiktok_category_path
WHERE category_id = ? AND shop_type <=> ? AND version <=> ? AND seller_type <=> ? AND deleted_at IS NULL`,
		categoryId,
		shopType,
		version,
		sellerType,
	)

	path := &erp_model.Temp1TiktokCategoryPath{}
	err := row.Scan(
		&path.Id,
		&path.CategoryId,
		&path.CategoryPathCn,
		&path.CategoryPathId,
		&path.CategoryPathEn,
		&path.ShopType,
		&path.Version,
		&path.SellerType,
		&path.Status,
		&path.CreatedAt,
		&path.UpdatedAt,
		&path.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category_path table: ", err.Error())
		return nil, err
	}
	return path, nil
}

func (*temp1TiktokCategoryPathDao) ListForAttrSync(session *dbsession.DBSession, filter TiktokCategoryPathAttrSyncFilter) ([]*erp_model.Temp1TiktokCategoryPath, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 500
	}

	fromSQL := "FROM temp1_tiktok_category_path p"
	whereSQL := []string{
		"p.deleted_at IS NULL",
		"p.category_id <> ''",
		"p.shop_type IN (1, 2)",
	}
	args := make([]interface{}, 0)
	if filter.Status >= 0 {
		whereSQL = append(whereSQL, "p.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Version != "" {
		if strings.EqualFold(filter.Version, "null") {
			whereSQL = append(whereSQL, "p.version IS NULL")
		} else {
			whereSQL = append(whereSQL, "p.version = ?")
			args = append(args, filter.Version)
		}
	}
	if filter.SellerType != "" {
		if strings.EqualFold(filter.SellerType, "null") {
			whereSQL = append(whereSQL, "p.seller_type IS NULL")
		} else {
			whereSQL = append(whereSQL, "p.seller_type = ?")
			args = append(args, filter.SellerType)
		}
	}

	if filter.SkipExisting {
		fromSQL += ` LEFT JOIN temp1_tiktok_category_attr a
ON a.category_id = p.category_id
AND a.shop_type <=> p.shop_type
AND a.version <=> p.version
AND a.seller_type <=> p.seller_type
AND a.deleted_at IS NULL`
		whereSQL = append(whereSQL, "a.id IS NULL")
	}
	if filter.ShopType > 0 {
		whereSQL = append(whereSQL, "p.shop_type = ?")
		args = append(args, filter.ShopType)
	}
	if filter.AfterID > 0 {
		whereSQL = append(whereSQL, "p.id > ?")
		args = append(args, filter.AfterID)
	}

	args = append(args, limit)
	query := `SELECT p.id, p.category_id, p.category_path_cn, p.category_path_id, p.category_path_en, p.shop_type, p.version, p.seller_type, p.status, p.created_at, p.updated_at, p.deleted_at
` + fromSQL + `
WHERE ` + strings.Join(whereSQL, " AND ") + `
ORDER BY p.id ASC
LIMIT ?`

	rows, err := session.Query(query, args...)
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category_path table: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	paths := make([]*erp_model.Temp1TiktokCategoryPath, 0)
	for rows.Next() {
		path := &erp_model.Temp1TiktokCategoryPath{}
		err := rows.Scan(
			&path.Id,
			&path.CategoryId,
			&path.CategoryPathCn,
			&path.CategoryPathId,
			&path.CategoryPathEn,
			&path.ShopType,
			&path.Version,
			&path.SellerType,
			&path.Status,
			&path.CreatedAt,
			&path.UpdatedAt,
			&path.DeletedAt,
		)
		if err != nil {
			log.Error("Failed to scan temp1_tiktok_category_path rows: ", err.Error())
			return nil, err
		}
		paths = append(paths, path)
	}
	if err := rows.Err(); err != nil {
		log.Error("Failed to iterate temp1_tiktok_category_path rows: ", err.Error())
		return nil, err
	}

	return paths, nil
}

func (*temp1TiktokCategoryPathDao) UpdateStatusByID(session *dbsession.DBSession, id uint64, status int) error {
	_, err := session.Exec(
		`UPDATE temp1_tiktok_category_path SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status,
		id,
	)
	if err != nil {
		log.Error("Failed to update temp1_tiktok_category_path status: ", err.Error())
	}
	return err
}
