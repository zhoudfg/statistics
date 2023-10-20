package viva_vn_api_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userDiamondRecordDao struct{}

var UserDiamondRecordDao = new(userDiamondRecordDao)

type userDiamondRecord struct{
	UserId int
	Type   int32
	Qty    int64
	Os     sql.NullString
}

// 用户钻石消耗记录
func (*userDiamondRecordDao) GetConsumeDiamondRecordByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userDiamondRecord {
	rows, err := session.Query("SELECT user_id, type, qty from user_diamond_record WHERE action = 0 AND create_time >= ? AND create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user_diamond_record table: ", err.Error())
		return nil
	}
	defer rows.Close()

	consumeRecord := make([]*userDiamondRecord, 0)
	for rows.Next() {
		consume := &userDiamondRecord{}
		err := rows.Scan(&consume.UserId, &consume.Type, &consume.Qty)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		consumeRecord = append(consumeRecord, consume)
	}

	return consumeRecord
}

// 消耗钻石类型数
func (*userDiamondRecordDao) GetConsumeDiamondTypeQtyByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userDiamondRecord {
	rows, err := session.Query("SELECT type, qty, user_register_info.mobile_os AS os from user_diamond_record LEFT JOIN user_register_info ON user_diamond_record.user_id = user_register_info.user_id WHERE action = 0 AND create_time >= ? AND create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	consumeDiamond := make([]*userDiamondRecord, 0)
	for rows.Next() {
		consume := &userDiamondRecord{}
		err := rows.Scan(&consume.Type, &consume.Qty, &consume.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		consumeDiamond = append(consumeDiamond, consume)
	}

	return consumeDiamond
}

// 产出钻石类型数
func (*userDiamondRecordDao) GetOutputDiamondTypeQtyByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userDiamondRecord {
	rows, err := session.Query("SELECT type, qty, user_register_info.mobile_os AS os from user_diamond_record LEFT JOIN user_register_info ON user_diamond_record.user_id = user_register_info.user_id where action = 1 AND create_time >= ? AND create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	outputDiamond := make([]*userDiamondRecord, 0)
	for rows.Next() {
		consume := &userDiamondRecord{}
		err := rows.Scan(&consume.Type, &consume.Qty, &consume.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		outputDiamond = append(outputDiamond, consume)
	}

	return outputDiamond
}
