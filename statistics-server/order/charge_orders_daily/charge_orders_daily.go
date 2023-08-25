package charge_orders_daily

import (
	"bytes"
	"fmt"
	"time"

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
	timeObj := time.Unix(int64(startTime), 0)
	date := timeObj.Format("2006-01-02")

	now := time.Now().Unix()
	if now < endTime {
		endTime = now - 5
	}

	{
		// 每天订单
		chargeOrderStatus := viva_vn_api.ChargeOrdersDao.GetChargeOrdersStatusByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var createOrderAndroidQty, completeOrderAndroidQty, uncompleteOrderAndroidQty, createOrderIosQty, completeOrderIosQty, uncompleteOrderIosQty int64
	
		for _, item := range chargeOrderStatus {
			if item.Status == 8 {
				if item.Os.String == "android" {
					completeOrderAndroidQty += 1
				}else if item.Os.String == "ios" {
					completeOrderIosQty += 1
				}
			}else{
				if item.Os.String == "android" {
					uncompleteOrderAndroidQty += 1
				}else if item.Os.String == "ios" {
					uncompleteOrderIosQty += 1
				}
			}
		}

		createOrderAndroidQty = completeOrderAndroidQty + uncompleteOrderAndroidQty
		createOrderIosQty = completeOrderIosQty + uncompleteOrderIosQty

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
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 6, uncompleteOrderIosQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeOrdersDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_orders_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		createOrderAndroidQty, completeOrderAndroidQty, uncompleteOrderAndroidQty, createOrderIosQty, completeOrderIosQty, uncompleteOrderIosQty)	
	}

	

}
