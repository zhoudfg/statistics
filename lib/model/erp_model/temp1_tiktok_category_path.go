package erp_model

import (
	"database/sql"
	"time"
)

type Temp1TiktokCategoryPath struct {
	Id             uint64         `json:"id" db:"id"`
	CategoryId     string         `json:"category_id" db:"category_id"`
	CategoryPathCn string         `json:"category_path_cn" db:"category_path_cn"`
	CategoryPathId string         `json:"category_path_id" db:"category_path_id"`
	CategoryPathEn string         `json:"category_path_en" db:"category_path_en"`
	ShopType       sql.NullInt64  `json:"shop_type" db:"shop_type"`
	Version        sql.NullString `json:"version" db:"version"`
	SellerType     sql.NullString `json:"seller_type" db:"seller_type"`
	Status         int            `json:"status" db:"status"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt      sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}
