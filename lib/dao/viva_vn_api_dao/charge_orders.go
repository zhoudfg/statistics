package viva_vn_api_dao

import (

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type chargeOrdersDao struct{}

var ChargeOrdersDao = new(chargeOrdersDao)

type chargeChannel struct{
	Diamond int64
	Channel string
}

type chargeOrdersStatus struct{
	Status int
}

type chargeOrdersRange struct{
	UserId  int64
	Diamond int64
}

type chargeOrdersDiamond struct{
	Diamond int64
}

// 获取充值渠道
func (*chargeOrdersDao) GetChargeChannelByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeChannel {
	rows, err := session.Query("SELECT charge_orders.diamond as diamond, charge_rules.channel as channel from charge_orders left join charge_rules on charge_orders.product_id = charge_rules.product_id where complete_time >= ? and complete_time < ?", startTime, endTime)

	var chargeChannels = make([]*chargeChannel, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
        channel := &chargeChannel{}
		err := rows.Scan(&channel.Diamond, &channel.Channel)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		chargeChannels = append(chargeChannels, channel)
	}

	return chargeChannels
}

// 获取充值订单
func (*chargeOrdersDao) GetChargeOrdersStatusByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrdersStatus {
	rows, err := session.Query("SELECT status from charge_orders where create_time >= ? and create_time < ?", startTime, endTime)

	var ordersStatus = make([]*chargeOrdersStatus, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
        status := &chargeOrdersStatus{}
		err := rows.Scan(&status.Status)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersStatus = append(ordersStatus, status)
	}

	return ordersStatus
}

// 获取充值订单钻石范围
func (*chargeOrdersDao) GetChargeOrdersRangeByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrdersRange {
	rows, err := session.Query("SELECT user_id, diamond from charge_orders where complete_time >= ? and complete_time < ?", startTime, endTime)

	var ordersRange = make([]*chargeOrdersRange, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
        orderRange := &chargeOrdersRange{}
		err := rows.Scan(&orderRange.UserId, &orderRange.Diamond)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersRange = append(ordersRange, orderRange)
	}

	return ordersRange
}

// 获取充值订单钻石档位
func (*chargeOrdersDao) GetChargeOrdersDiamondByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrdersDiamond {
	rows, err := session.Query("SELECT diamond from charge_orders where complete_time >= ? and complete_time < ?", startTime, endTime)

	var ordersDiamond = make([]*chargeOrdersDiamond, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
        diamond := &chargeOrdersDiamond{}
		err := rows.Scan(&diamond.Diamond)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersDiamond = append(ordersDiamond, diamond)
	}

	return ordersDiamond
}
