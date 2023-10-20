package viva_vn_api_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type userGiftConsumeRecordDao struct{}

var UserGiftConsumeRecordDao = new(userGiftConsumeRecordDao)

type userGiftConsumeRecord struct{
	UserId        int
	ReceiveUserId int
	ScenesType    int
	ScenesId      int
	GiftId        int
	GiftQty       int
	GiftOriginalPrice  int
	GiftPrice     int
	GiftValue     int
	GiftWithdrawalValue  int
}

// 获取送作品钻石礼物消耗记录
func (*userGiftConsumeRecordDao) GetSongSendGiftByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*userGiftConsumeRecord {
	rows, err := session.Query("SELECT user_id, gift_qty, gift_price from user_gift_consume_record where scenes_type=2 and type=1 and create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user_gift_consume_record table: ", err.Error())
		return nil
	}
	defer rows.Close()

	userGiftConsumes := make([]*userGiftConsumeRecord, 0)
	for rows.Next() {
		consume := &userGiftConsumeRecord{}
		err := rows.Scan(&consume.UserId, &consume.GiftQty, &consume.GiftPrice)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		userGiftConsumes = append(userGiftConsumes, consume)
	}

	return userGiftConsumes
}

