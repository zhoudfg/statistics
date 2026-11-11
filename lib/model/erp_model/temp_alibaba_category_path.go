package erp_model

type TempAlibabaCategoryPath struct {
	Id           int64  `json:"id" db:"id"`
	CategoryType int    `json:"category_type" db:"category_type"`
	CategoryId   int64  `json:"category_id" db:"category_id"`
	Scene        string `json:"scene" db:"scene"`
}
