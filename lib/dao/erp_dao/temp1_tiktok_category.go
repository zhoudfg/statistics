package erp_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

type temp1TiktokCategoryDao struct{}

var Temp1TiktokCategoryDao = new(temp1TiktokCategoryDao)

func (*temp1TiktokCategoryDao) Insert(session *dbsession.DBSession, category *erp_model.Temp1TiktokCategory) (int64, error) {
	result, err := session.Exec(
		`INSERT INTO temp1_tiktok_category
(category_id, parent_id, is_leaf, name_cn, name_en, permission_statuses, shop_type, status, seller_type, version, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		category.CategoryId,
		category.ParentId,
		category.IsLeaf,
		category.NameCn,
		category.NameEn,
		[]byte(category.PermissionStatuses),
		category.ShopType,
		category.Status,
		category.SellerType,
		category.Version,
		category.DeletedAt,
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

func (*temp1TiktokCategoryDao) Upsert(session *dbsession.DBSession, category *erp_model.Temp1TiktokCategory) (int64, error) {
	result, err := session.Exec(
		`INSERT INTO temp1_tiktok_category
(category_id, parent_id, is_leaf, name_cn, name_en, permission_statuses, shop_type, status, seller_type, version, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE parent_id = VALUES(parent_id), is_leaf = VALUES(is_leaf), name_cn = VALUES(name_cn), name_en = VALUES(name_en), permission_statuses = VALUES(permission_statuses), status = VALUES(status), deleted_at = VALUES(deleted_at), updated_at = CURRENT_TIMESTAMP`,
		category.CategoryId,
		category.ParentId,
		category.IsLeaf,
		category.NameCn,
		category.NameEn,
		[]byte(category.PermissionStatuses),
		category.ShopType,
		category.Status,
		category.SellerType,
		category.Version,
		category.DeletedAt,
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

func (*temp1TiktokCategoryDao) GetByUniqueKey(session *dbsession.DBSession, categoryId string, version sql.NullString, sellerType string, shopType int) (*erp_model.Temp1TiktokCategory, error) {
	row := session.QueryRow(
		`SELECT id, category_id, parent_id, is_leaf, name_cn, name_en, permission_statuses, shop_type, status, seller_type, version, created_at, updated_at, deleted_at
FROM temp1_tiktok_category
WHERE category_id = ? AND version <=> ? AND seller_type = ? AND shop_type = ? AND deleted_at IS NULL`,
		categoryId,
		version,
		sellerType,
		shopType,
	)

	category := &erp_model.Temp1TiktokCategory{}
	var permissionStatuses []byte
	err := row.Scan(
		&category.Id,
		&category.CategoryId,
		&category.ParentId,
		&category.IsLeaf,
		&category.NameCn,
		&category.NameEn,
		&permissionStatuses,
		&category.ShopType,
		&category.Status,
		&category.SellerType,
		&category.Version,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category table: ", err.Error())
		return nil, err
	}
	category.PermissionStatuses = append(category.PermissionStatuses, permissionStatuses...)
	return category, nil
}

func (*temp1TiktokCategoryDao) GetChildren(session *dbsession.DBSession, parentId string, version sql.NullString, sellerType string, shopType int) ([]*erp_model.Temp1TiktokCategory, error) {
	rows, err := session.Query(
		`SELECT id, category_id, parent_id, is_leaf, name_cn, name_en, permission_statuses, shop_type, status, seller_type, version, created_at, updated_at, deleted_at
FROM temp1_tiktok_category
WHERE parent_id = ? AND version <=> ? AND seller_type = ? AND shop_type = ? AND deleted_at IS NULL`,
		parentId,
		version,
		sellerType,
		shopType,
	)
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category table: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	categories := make([]*erp_model.Temp1TiktokCategory, 0)
	for rows.Next() {
		category := &erp_model.Temp1TiktokCategory{}
		var permissionStatuses []byte
		err := rows.Scan(
			&category.Id,
			&category.CategoryId,
			&category.ParentId,
			&category.IsLeaf,
			&category.NameCn,
			&category.NameEn,
			&permissionStatuses,
			&category.ShopType,
			&category.Status,
			&category.SellerType,
			&category.Version,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil, err
		}
		category.PermissionStatuses = append(category.PermissionStatuses, permissionStatuses...)
		categories = append(categories, category)
	}
	return categories, nil
}
