package erp_model

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Temp1TiktokCategoryAttr struct {
	Id         uint64          `json:"id" db:"id"`
	CategoryId string          `json:"category_id" db:"category_id"`
	RawJson    json.RawMessage `json:"raw_json" db:"raw_json"`
	ShopType   sql.NullInt64   `json:"shop_type" db:"shop_type"`
	Version    sql.NullString  `json:"version" db:"version"`
	SellerType sql.NullString  `json:"seller_type" db:"seller_type"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt  sql.NullTime    `json:"deleted_at" db:"deleted_at"`
}

