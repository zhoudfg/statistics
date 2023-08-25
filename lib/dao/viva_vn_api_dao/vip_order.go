package viva_vn_api_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type vipOrderDao struct{}

var VipOrderDao = new(vipOrderDao)

type vipOrderDays struct{
	UserId int64
	Days   int64
	Os   sql.NullString
}

type vipOrder struct{
	Id int64
	UserId int64
	Os   sql.NullString
}
// 获取vip订单数
func (*vipOrderDao) GetVipOrderQtyByTime(session *dbsession.DBSession, startTime int64, endTime int64) int64 {
	row := session.QueryRow("SELECT count(id) as count from vip_order where create_time >= ? and create_time < ?", startTime, endTime)
	
	var count int64
	
	err := row.Scan(&count)
	if err != nil {
		log.Error("Failed to scan row: ", err)
		return 0
	}

	return count
}

// 获取vip订单数
func (*vipOrderDao) GetVipOrderByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*vipOrder {
	rows, err := session.Query("SELECT id, user_register_info.mobile_os as os from vip_order left join user_register_info on user_register_info.user_id = vip_order.user_id where create_time >= ? and create_time < ?", startTime, endTime)
	
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	vipOrders := make([]*vipOrder, 0)
	for rows.Next() {
		vip := &vipOrder{}
		err := rows.Scan(&vip.Id, &vip.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		vipOrders = append(vipOrders, vip)
	}

	return vipOrders
}

// 获取vip充值天数
func (*vipOrderDao) GetVipOrderDaysByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*vipOrderDays {
	rows, err := session.Query("SELECT vip_order.user_id AS user_id, vip_package.day as day, user_register_info.mobile_os AS os from vip_order left join vip_package on vip_order.package_id = vip_package.id left join user_register_info on user_register_info.user_id = vip_order.user_id where vip_order.create_time >= ? and vip_order.create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	orderDays := make([]*vipOrderDays, 0)
	for rows.Next() {
        days := &vipOrderDays{}
		err := rows.Scan(&days.UserId, &days.Days, &days.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		orderDays = append(orderDays, days)
	}

	return orderDays
}