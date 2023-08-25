package vip_orders_days_daily

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
		// 获取vip订单天数
		vipOrderDays := viva_vn_api.VipOrderDao.GetVipOrderDaysByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		userDaysMap := make(map[int64]int64, 0)
		userOsMap := make(map[int64]string, 0)

		for _, item := range vipOrderDays {
			userDaysMap[item.UserId] += item.Days
			userOsMap[item.UserId] = item.Os.String
		}

		var days30AndroidQty, day60AndroidQty, day180AndroidQty, day360AndroidQty,
			days30IosQty, day60IosQty, day180IosQty, day360IosQty int64

		for userId, days := range userDaysMap {
			os, ok := userOsMap[userId]
			if !ok {
				continue
			}

			if os == "android" {
				if days >= 30 && days < 60 {
					days30AndroidQty += 1
				} else if days >= 60 && days < 180 {
					day60AndroidQty += 1
				} else if days >= 180 && days < 360 {
					day180AndroidQty += 1
				} else if days >= 360 {
					day360AndroidQty += 1
				}
			} else if os == "ios" {
				if days >= 30 && days < 60 {
					days30IosQty += 1
				} else if days >= 60 && days < 180 {
					day60IosQty += 1
				} else if days >= 180 && days < 360 {
					day180IosQty += 1
				} else if days >= 360 {
					day360IosQty += 1
				}
			}
		}

		totalAndroidQty := days30AndroidQty + day60AndroidQty + day180AndroidQty + day360AndroidQty
		totalIosQty := days30IosQty + day60IosQty + day180IosQty + day360IosQty

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "vip_orders_days_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, days30AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, day60AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, day180AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, day360AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, totalAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, days30IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, day60IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, day180IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, day360IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 10, totalIosQty))

		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.VipOrdersDaysDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily vip_orders_days_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", days30AndroidQty, day60AndroidQty, day180AndroidQty, day360AndroidQty,
			totalAndroidQty, days30IosQty, day60IosQty, day180IosQty, day360IosQty, totalIosQty)
	}

}
