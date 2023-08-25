package viva_vn_api_dao

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userPersonInfoDao struct{}

var UserPersonInfoDao = new(userPersonInfoDao)

type userPersonInfo struct{
   UserId   int64
   Birthday int64
   Os       string
}
// 获取用户年龄信息
func (*userPersonInfoDao) GetActivityUserAgeByUserIds(session *dbsession.DBSession, userIds []int64) []*userPersonInfo {

	if len(userIds) == 0 {
		return nil
	}

	userIdsInterface := make([]interface{}, len(userIds))
	for i, id := range userIds {
		userIdsInterface[i] = id
	}

	placeholders := make([]string, len(userIds))
	for i := range userIds {
		placeholders[i] = "?"
	}
	
	query := fmt.Sprintf("SELECT user_person_info.user_id AS uid, birthday, user_register_info.mobile_os AS os FROM user_person_info LEFT JOIN user_register_info ON user_register_info.user_id = user_person_info.user_id WHERE user_id IN (%s)", strings.Join(placeholders, ", "))

	rows, err := session.Query(query, userIdsInterface...)
	
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	userAges := make([]*userPersonInfo, 0)
	for rows.Next() {
		user := &userPersonInfo{}
		err := rows.Scan(&user.UserId, &user.Birthday, &user.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		userAges = append(userAges, user)
	}

	return userAges
}


