package viva_vn_statistics_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"

	"github.com/statistics/lib/dbsession"
)

type userFirstChargeLogDao struct{}

var UserFirstChargeLogDao = new(userFirstChargeLogDao)

type userFirstChargeLog struct{
	UserId    int
	Price     float64
	PriceUnit sql.NullString
	Diamond   int
	ProductId sql.NullString
	day       int
}

func (*userFirstChargeLogDao) BatchInsertOrUpdate(session *dbsession.DBSession, sql string) (int, error) {
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

func (*userFirstChargeLogDao) GetUserFirstChargeUserAmountByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userFirstChargeLog {
	rows, err := session.Query("SELECT user_id, price, price_unit from user_first_charge_log where day >= ? and day < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user_first_charge_log table: ", err.Error())
		return nil
	}
	defer rows.Close()

	userFirstCharges := make([]*userFirstChargeLog, 0)

	for rows.Next() {
		userFirstCharge := &userFirstChargeLog{}
		err := rows.Scan(&userFirstCharge.UserId, &userFirstCharge.Price, &userFirstCharge.PriceUnit)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		userFirstCharges = append(userFirstCharges, userFirstCharge)
	}

	return userFirstCharges

}

