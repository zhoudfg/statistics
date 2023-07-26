package viva_vn_api

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userDao struct{}

var UserDao = new(userDao)

// 获取当天注册用户
func (*userDao) GetNewUserByTime(session *dbsession.DBSession, startTime int64, endTime int64) []interface{} {
	rows, err := session.Query("SELECT id from user where create_time >= ? and create_time < ?", startTime, endTime)
	var Id int64
	var user = make([]interface{},0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return user
	}
	for rows.Next() {
		err := rows.Scan(&Id)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return user
		}
		user = append(user, Id)
	}

	return user
}





