package viva_vn_statistics_model

const (
	AndroidOldUser int = iota + 1
	AndroidNewUser
	AndroidTotalUser
	IosOldUser
	IosNewUser
	IosTotalUser
)

type UserActivityDailyLog struct {
	Day   string `json:"day"`
	Type  int    `json:"type"`
	Qty   int    `json:"qty"`
}
