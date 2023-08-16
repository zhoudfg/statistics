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
	timeStr := time.Now().Format("2006-01-02")
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

        for _, item := range vipOrderDays{
			userDaysMap[item.UserId] += item.Days
		}

		var days30Qty, day60Qty, day180Qty, day360Qty, totalQty int64

		for _, days := range userDaysMap{
			if days >= 30 && days < 60{
				days30Qty += 1
			}else if days >= 60 && days < 180{
				day60Qty += 1
			}else if days >= 180 && days < 360{
				day180Qty += 1
			}else if days >= 360{
				day360Qty += 1
			}
		}

		totalQty = days30Qty + day60Qty + day180Qty + day360Qty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "vip_orders_days_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, days30Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, day60Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, day180Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, day360Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 5, totalQty))

		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.VipOrdersDaysDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily vip_orders_days_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", days30Qty,
		day60Qty, day180Qty, day360Qty, totalQty)	
	}

	

}
