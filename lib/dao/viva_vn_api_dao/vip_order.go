package viva_vn_api_dao

import (

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type vipOrderDao struct{}

var VipOrderDao = new(vipOrderDao)

type vipOrderDays struct{
	UserId int64
	Days   int64
}
// 获取vip订单数
func (*vipOrderDao) GetVipOrderByTime(session *dbsession.DBSession, startTime int64, endTime int64) int64 {
	row := session.QueryRow("SELECT count(id) as count from vip_order where create_time >= ? and create_time < ?", startTime, endTime)
	
	var count int64

	err := row.Scan(&count)
	if err != nil {
		log.Error("Failed to scan row: ", err)
		return 0
	}

	return count
}

// 获取vip充值天数
func (*vipOrderDao) GetVipOrderDaysByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*vipOrderDays {
	rows, err := session.Query("SELECT user_id, vip_package.day as day from vip_order left join vip_package on vip_order.package_id = vip_package.id where vip_order.create_time >= ? and vip_order.create_time < ?", startTime, endTime)

	var orderDays = make([]*vipOrderDays, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
        days := &vipOrderDays{}
		err := rows.Scan(&days.UserId, &days.Days)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		orderDays = append(orderDays, days)
	}

	return orderDays
}