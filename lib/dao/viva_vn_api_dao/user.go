package viva_vn_api_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userDao struct{}

var UserDao = new(userDao)

type userInfo struct {
	Id  int
	Sex sql.NullInt64
	Os  string
}

// 获取当天注册用户
func (*userDao) GetNewUserByTime(session *dbsession.DBSession, startTime int64, endTime int64) []interface{} {
	rows, err := session.Query("SELECT id from user where create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	var Id int64
	user := make([]interface{}, 0)
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

// 每天注册用户性别
func (*userDao) GetRegisterUserSexByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userInfo {
	rows, err := session.Query("SELECT id, user_person_info.sex as sex, user_register_info.mobile_os as os from user left join user_person_info on user.id = user_person_info.user_id left join user_register_info on user_register_info.user_id = user.id where create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	users := make([]*userInfo, 0)
	for rows.Next() {
		user := &userInfo{}
		err := rows.Scan(&user.Id, &user.Sex)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		users = append(users, user)
	}

	return users
}
