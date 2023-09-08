package vip_orders_daily

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
	timeObj := time.Unix(int64(startTime)+(8*3600), 0).UTC()
	date := timeObj.Format("2006-01-02")

	now := time.Now().Unix()
	if now < endTime {
		endTime = now - 5
	}

	{
		// 获取vip订单数
		vipOrders := viva_vn_api.VipOrderDao.GetVipOrderByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var androidQty, iosQty int64

		for _, item := range vipOrders {
			if item.Os.String == "android" {
				androidQty += 1
			} else if item.Os.String == "ios" {
				iosQty += 1
			}
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "vip_orders_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, androidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 2, iosQty))

		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.VipOrdersDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily vip_orders_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", androidQty, iosQty)
	}

}
