package charge_orders_daily

import (
	"bytes"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"
	viva_vn_api "github.com/statistics/lib/dao/viva_vn_api_dao"
	"github.com/statistics/lib/dao/viva_vn_statistics_dao"
	"github.com/statistics/lib/db"
)

// day表示几天前,0为当日
func CountDaily(day int64) {
	timeStr := time.Now().In(time.Local).Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timeStr)
	startTime := t.Unix() - (8 * 3600) - 86400*day
	endTime := t.Unix() - (8 * 3600) - (86400 * (day - 1))
	timeObj := time.Unix(int64(startTime)+(8*3600), 0).UTC()
	date := timeObj.Format("2006-01-02")

	now := time.Now().Unix()
	if now < endTime {
		endTime = now - 5
	}

	{
		// 每天订单
		chargeOrderStatus := viva_vn_api.ChargeOrdersDao.GetChargeOrdersStatusByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		// 当天注册的新用户
		user := viva_vn_api.UserDao.GetNewUserByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var set = mapset.NewSetFromSlice(user)

		var createOrderAndroidQty, completeOrderAndroidQty, uncompleteOrderAndroidQty, createOrderIosQty, completeOrderIosQty, uncompleteOrderIosQty,
		    createOrderNewUserQty, completeOrderNewUserQty, uncompleteOrderNewUserQty, createOrderOldUserQty, completeOrderOldUserQty, uncompleteOrderOldUserQty int64
	
		for _, item := range chargeOrderStatus {
			if item.Status == 8 {
				if item.Os.String == "android" {
					completeOrderAndroidQty += 1
				}else if item.Os.String == "ios" {
					completeOrderIosQty += 1
				}
				if set.Contains(item.UserId) {
					completeOrderNewUserQty += 1
				}else{
					completeOrderOldUserQty += 1
				}
			}else{
				if item.Os.String == "android" {
					uncompleteOrderAndroidQty += 1
				}else if item.Os.String == "ios" {
					uncompleteOrderIosQty += 1
				}
				if set.Contains(item.UserId) {
					uncompleteOrderNewUserQty += 1
				}else{
					uncompleteOrderOldUserQty += 1
				}
			}
		}

		createOrderAndroidQty = completeOrderAndroidQty + uncompleteOrderAndroidQty
		createOrderIosQty = completeOrderIosQty + uncompleteOrderIosQty

		createOrderNewUserQty = completeOrderNewUserQty + uncompleteOrderNewUserQty
		createOrderOldUserQty = completeOrderOldUserQty + uncompleteOrderOldUserQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_orders_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, createOrderAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, completeOrderAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, uncompleteOrderAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, createOrderIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, completeOrderIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, uncompleteOrderIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, createOrderNewUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, completeOrderNewUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, uncompleteOrderNewUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 10, createOrderOldUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 11, completeOrderOldUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 12, uncompleteOrderOldUserQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeOrdersDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_orders_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		createOrderAndroidQty, completeOrderAndroidQty, uncompleteOrderAndroidQty, createOrderIosQty, completeOrderIosQty, uncompleteOrderIosQty,
		createOrderNewUserQty, completeOrderNewUserQty, uncompleteOrderNewUserQty, createOrderOldUserQty, completeOrderOldUserQty, uncompleteOrderOldUserQty)	
	}

	

}
