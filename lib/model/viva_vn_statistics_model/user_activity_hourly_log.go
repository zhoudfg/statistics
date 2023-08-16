package viva_vn_statistics_model

const (
	AndroidUser int = iota + 1
	IosUser
	TotalUser
)

type UserActivityHourlyLog struct {
	Day   string `json:"day"`
	Type  int    `json:"type"`
	Qty   int    `json:"qty"`
}
