package erp_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

type temp1TiktokCategoryAttrDao struct{}

var Temp1TiktokCategoryAttrDao = new(temp1TiktokCategoryAttrDao)

func (*temp1TiktokCategoryAttrDao) Insert(session *dbsession.DBSession, attr *erp_model.Temp1TiktokCategoryAttr) (int64, error) {
	result, err := session.Exec(
		"INSERT INTO temp1_tiktok_category_attr (category_id, raw_json, shop_type, version, seller_type, deleted_at) VALUES (?, ?, ?, ?, ?, ?)",
		attr.CategoryId,
		[]byte(attr.RawJson),
		attr.ShopType,
		attr.Version,
		attr.SellerType,
		attr.DeletedAt,
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

func (*temp1TiktokCategoryAttrDao) Upsert(session *dbsession.DBSession, attr *erp_model.Temp1TiktokCategoryAttr) (int64, error) {
	result, err := session.Exec(
		`INSERT INTO temp1_tiktok_category_attr (category_id, raw_json, shop_type, version, seller_type, deleted_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE raw_json = VALUES(raw_json), deleted_at = VALUES(deleted_at), updated_at = CURRENT_TIMESTAMP`,
		attr.CategoryId,
		[]byte(attr.RawJson),
		attr.ShopType,
		attr.Version,
		attr.SellerType,
		attr.DeletedAt,
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

func (*temp1TiktokCategoryAttrDao) GetByUniqueKey(session *dbsession.DBSession, categoryId string, shopType sql.NullInt64, version sql.NullString, sellerType sql.NullString) (*erp_model.Temp1TiktokCategoryAttr, error) {
	row := session.QueryRow(
		`SELECT id, category_id, raw_json, shop_type, version, seller_type, created_at, updated_at, deleted_at
FROM temp1_tiktok_category_attr
WHERE category_id = ? AND shop_type <=> ? AND version <=> ? AND seller_type <=> ? AND deleted_at IS NULL`,
		categoryId,
		shopType,
		version,
		sellerType,
	)

	attr := &erp_model.Temp1TiktokCategoryAttr{}
	var rawJson []byte
	err := row.Scan(&attr.Id, &attr.CategoryId, &rawJson, &attr.ShopType, &attr.Version, &attr.SellerType, &attr.CreatedAt, &attr.UpdatedAt, &attr.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category_attr table: ", err.Error())
		return nil, err
	}
	attr.RawJson = append(attr.RawJson, rawJson...)
	return attr, nil
}
