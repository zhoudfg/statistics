package erp_model

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Temp1TiktokCategory struct {
	Id                 int             `json:"id" db:"id"`
	CategoryId         string          `json:"category_id" db:"category_id"`
	ParentId           string          `json:"parent_id" db:"parent_id"`
	IsLeaf             int             `json:"is_leaf" db:"is_leaf"`
	NameCn             string          `json:"name_cn" db:"name_cn"`
	NameEn             string          `json:"name_en" db:"name_en"`
	PermissionStatuses json.RawMessage `json:"permission_statuses" db:"permission_statuses"`
	ShopType           int             `json:"shop_type" db:"shop_type"`
	Status             sql.NullString  `json:"status" db:"status"`
	SellerType         string          `json:"seller_type" db:"seller_type"`
	Version            sql.NullString  `json:"version" db:"version"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt          sql.NullTime    `json:"deleted_at" db:"deleted_at"`
}

