package viva_vn_api_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userDiamondRecordDao struct{}

var UserDiamondRecordDao = new(userDiamondRecordDao)

type diamondType struct{
	Type int32
	Qty int64
}

// 消耗钻石类型数
func (*userDiamondRecordDao) GetConsumeDiamondTypeQtyByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*diamondType {
	rows, err := session.Query("SELECT type, qty from user_diamond_record where action = 0 AND create_time >= ? AND create_time < ?", startTime, endTime)

	var consumeDiamond = make([]*diamondType, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
		consume := &diamondType{}
		err := rows.Scan(&consume.Type, &consume.Qty)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		consumeDiamond = append(consumeDiamond, consume)
	}

	return consumeDiamond
}

// 消耗钻石类型数
func (*userDiamondRecordDao) GetOutputDiamondTypeQtyByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*diamondType {
	rows, err := session.Query("SELECT type, qty from user_diamond_record where action = 1 AND create_time >= ? AND create_time < ?", startTime, endTime)

	var outputDiamond = make([]*diamondType, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
		consume := &diamondType{}
		err := rows.Scan(&consume.Type, &consume.Qty)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		outputDiamond = append(outputDiamond, consume)
	}

	return outputDiamond
}
