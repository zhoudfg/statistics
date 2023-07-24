package model

//用户登录纪录
type UserLoginRecord struct {
	AppId       string `bson:"app_id" json:"app_id"`
	UserId      int64  `bson:"user_id" json:"user_id"`
	LoginTime   int64  `bson:"login_time" json:"login_time"`
	LogoutTime  int64  `bson:"logout_time" json:"logout_time"`
	Uuid        string `bson:"uuid" json:"uuid"`
	Os          string `bson:"os" json:"os"`
	MobileOs    string `bson:"mobile_os" json:"mobile_os"`
	MobileModel string `bson:"mobile_model" json:"mobile_model"`
	VersionCode string `bson:"version_code" json:"version_code"`
	VersionName string `bson:"version_name" json:"version_name"`
}
