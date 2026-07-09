package erp_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

type temp1TiktokCategoryAttrValuesDao struct{}

var Temp1TiktokCategoryAttrValuesDao = new(temp1TiktokCategoryAttrValuesDao)

func (*temp1TiktokCategoryAttrValuesDao) Insert(session *dbsession.DBSession, value *erp_model.Temp1TiktokCategoryAttrValue) (int64, error) {
	result, err := session.Exec(
		"INSERT INTO temp1_tiktok_category_attr_values (value_id, value_name, deleted_at) VALUES (?, ?, ?)",
		value.ValueId,
		value.ValueName,
		value.DeletedAt,
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

func (*temp1TiktokCategoryAttrValuesDao) GetByValueId(session *dbsession.DBSession, valueId string) ([]*erp_model.Temp1TiktokCategoryAttrValue, error) {
	rows, err := session.Query(
		`SELECT id, value_id, value_name, created_at, updated_at, deleted_at
FROM temp1_tiktok_category_attr_values
WHERE value_id = ? AND deleted_at IS NULL`,
		valueId,
	)
	if err != nil {
		log.Error("Failed to query the temp1_tiktok_category_attr_values table: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	values := make([]*erp_model.Temp1TiktokCategoryAttrValue, 0)
	for rows.Next() {
		value := &erp_model.Temp1TiktokCategoryAttrValue{}
		err := rows.Scan(&value.Id, &value.ValueId, &value.ValueName, &value.CreatedAt, &value.UpdatedAt, &value.DeletedAt)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
