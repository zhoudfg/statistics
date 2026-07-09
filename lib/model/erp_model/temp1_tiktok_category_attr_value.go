package erp_model

import "database/sql"

type Temp1TiktokCategoryAttrValue struct {
	Id        int          `json:"id" db:"id"`
	ValueId   string       `json:"value_id" db:"value_id"`
	ValueName string       `json:"value_name" db:"value_name"`
	CreatedAt sql.NullTime `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

