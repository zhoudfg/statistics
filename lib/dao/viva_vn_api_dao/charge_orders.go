package viva_vn_api_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type chargeOrdersDao struct{}

var ChargeOrdersDao = new(chargeOrdersDao)

type chargeOrders struct{
	UserId     int64
	Diamond    int64
	Channel    sql.NullString
	ProductId  sql.NullString
	Status     int
	Os         sql.NullString
	Price      float64
	PriceUnit  sql.NullString
	ChargeType  sql.NullString
	CreateTime int
}

// 获取充值渠道
func (*chargeOrdersDao) GetChargeChannelByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT charge_orders.diamond AS diamond, charge_orders.product_id AS product_id, user_register_info.mobile_os AS os, charge_orders.charge_type AS charge_type FROM charge_orders LEFT JOIN user_register_info ON charge_orders.user_id = user_register_info.user_id WHERE complete_time >= ? and complete_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	chargeChannels := make([]*chargeOrders, 0)
	for rows.Next() {
        channel := &chargeOrders{}
		err := rows.Scan(&channel.Diamond, &channel.ProductId, &channel.Os, &channel.ChargeType)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		chargeChannels = append(chargeChannels, channel)
	}

	return chargeChannels
}

// 获取充值渠道2
func (*chargeOrdersDao) GetChargeChannel2ByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT user_id, charge_rules.channel AS channel, status FROM charge_orders LEFT JOIN charge_rules ON charge_rules.product_id = charge_orders.product_id WHERE create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	orders := make([]*chargeOrders, 0)
	for rows.Next() {
        chargeOrder := &chargeOrders{}
		err := rows.Scan(&chargeOrder.UserId, &chargeOrder.Channel, &chargeOrder.Status)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		orders = append(orders, chargeOrder)
	}

	return orders
}

// 获取充值订单1
func (*chargeOrdersDao) GetChargeOrders1ByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT status, user_register_info.mobile_os AS os FROM charge_orders LEFT JOIN user_register_info ON charge_orders.user_id = user_register_info.user_id WHERE create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	ordersStatus := make([]*chargeOrders, 0)
	for rows.Next() {
        status := &chargeOrders{}
		err := rows.Scan(&status.Status, &status.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersStatus = append(ordersStatus, status)
	}

	return ordersStatus
}

// 获取充值订单2
func (*chargeOrdersDao) GetChargeOrders2ByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT user_id, charge_orders.diamond AS diamond, charge_orders.product_id AS product_id, status, charge_orders.price AS price, charge_orders.price_unit AS price_unit, charge_rules.channel AS channel, charge_orders.charge_type AS charge_type FROM charge_orders LEFT JOIN charge_rules on charge_rules.product_id=charge_orders.product_id WHERE create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	orders := make([]*chargeOrders, 0)
	for rows.Next() {
        chargeOrder := &chargeOrders{}
		err := rows.Scan(&chargeOrder.UserId, &chargeOrder.Diamond, &chargeOrder.ProductId, &chargeOrder.Status, &chargeOrder.Price, &chargeOrder.PriceUnit, &chargeOrder.Channel, &chargeOrder.ChargeType)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		orders = append(orders, chargeOrder)
	}

	return orders
}

// 获取充值订单钻石范围
func (*chargeOrdersDao) GetChargeOrdersRangeByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT charge_orders.user_id AS user_id, diamond, user_register_info.mobile_os AS os FROM charge_orders LEFT JOIN user_register_info ON charge_orders.user_id = user_register_info.user_id WHERE complete_time >= ? AND complete_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	ordersRange := make([]*chargeOrders, 0)
	for rows.Next() {
        orderRange := &chargeOrders{}
		err := rows.Scan(&orderRange.UserId, &orderRange.Diamond, &orderRange.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersRange = append(ordersRange, orderRange)
	}

	return ordersRange
}

// 获取充值订单钻石档位
func (*chargeOrdersDao) GetChargeOrdersDiamondByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT diamond, user_register_info.mobile_os AS os FROM charge_orders LEFT JOIN user_register_info ON charge_orders.user_id = user_register_info.user_id WHERE complete_time >= ? AND complete_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	ordersDiamond := make([]*chargeOrders, 0)
	for rows.Next() {
        diamond := &chargeOrders{}
		err := rows.Scan(&diamond.Diamond, &diamond.Os)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		ordersDiamond = append(ordersDiamond, diamond)
	}

	return ordersDiamond
}

// 获取用户历史首充日期
func (*chargeOrdersDao) GetFirstChargeHistoryDate(session *dbsession.DBSession, startTime int64, endTime int64) []*chargeOrders {
	rows, err := session.Query("SELECT user_id, create_time, price, price_unit, diamond, product_id FROM charge_orders WHERE status = 8 AND create_time >= ? AND create_time < ?;", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the charge_orders table: ", err.Error())
		return nil
	}
	defer rows.Close()

	userIdDate := make([]*chargeOrders, 0)
	for rows.Next() {
         
        chargeOrders := &chargeOrders{}

		err := rows.Scan(&chargeOrders.UserId, &chargeOrders.CreateTime, &chargeOrders.Price, &chargeOrders.PriceUnit, &chargeOrders.Diamond, &chargeOrders.ProductId)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		userIdDate = append(userIdDate, chargeOrders)
	}

	return userIdDate
}