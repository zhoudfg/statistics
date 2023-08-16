package charge_orders_diamond_daily

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
		
		chargeOrderDiamond := viva_vn_api.ChargeOrdersDao.GetChargeOrdersDiamondByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var diamond1400Qty, diamond7000Qty, diamond14000Qty, diamond100000Qty, diamond180000Qty, diamond360000Qty, totalQty int64
	
		for _, item := range chargeOrderDiamond {
			if item.Diamond >= 1400 && item.Diamond < 7000 {
				diamond1400Qty += 1
			}else if item.Diamond >= 7000 && item.Diamond < 14000{
				diamond7000Qty += 1
			}else if item.Diamond >= 14000 && item.Diamond < 100000{
				diamond7000Qty += 1
			}else if item.Diamond >= 100000 && item.Diamond < 180000{
				diamond7000Qty += 1
			}else if item.Diamond >= 180000 && item.Diamond < 360000{
				diamond7000Qty += 1
			}else if item.Diamond >= 360000{
				diamond360000Qty += 1
			}
		}

		totalQty = diamond1400Qty + diamond7000Qty + diamond14000Qty + diamond100000Qty + diamond180000Qty + diamond360000Qty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_orders_diamond_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, diamond1400Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, diamond7000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, diamond14000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, diamond100000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, diamond180000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, diamond360000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 7, totalQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeOrdersDiamondDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_orders_diamond_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
			diamond1400Qty, diamond7000Qty, diamond14000Qty, diamond100000Qty, diamond180000Qty, diamond360000Qty, totalQty)	
	}

	

}
