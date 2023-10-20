package viva_vn_statistics_dao

import (
	log "github.com/sirupsen/logrus"

	"github.com/statistics/lib/dbsession"
)

type userFirstOpenDailyLogDao struct{}

var UserFirstOpenDailyLogDao = new(userFirstOpenDailyLogDao)

func (*userFirstOpenDailyLogDao) BatchInsertOrUpdate(session *dbsession.DBSession, sql string) (int, error) {
	result, err := session.Exec(sql)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return int(id), nil
}

func (*userFirstOpenDailyLogDao) GetUserFirstChargeUserAmountByTime(session *dbsession.DBSession, startTime int64, endTime int64) map[int]float64 {
	rows, err := session.Query("SELECT user_id, qty from user_first_open_daily_log where type=1 and day >= ? and day < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user_first_open_daily_log table: ", err.Error())
		return nil
	}
	defer rows.Close()

	var userId int
	var amount float64
	userIdAmountMap := make(map[int]float64, 0)

	for rows.Next() {

		err := rows.Scan(&userId, &amount)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		userIdAmountMap[userId] = amount
	}

	return userIdAmountMap

}

