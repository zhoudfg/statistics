package charge_orders_range_daily

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
		
		chargeOrderRange := viva_vn_api.ChargeOrdersDao.GetChargeOrdersRangeByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var diamond14000Qty, diamond14000t180000Qty, diamond180000t360000Qty, diamond360000Qty int64
	
		userDiamondMap := make(map[int64]int64, 0)
		for _, item := range chargeOrderRange {
			userDiamondMap[item.UserId] += item.Diamond
		}

		for _, diamond := range userDiamondMap {
			if diamond <= 14000 {
				diamond14000Qty += 1
			}else if diamond > 14000 && diamond <= 180000{
				diamond14000t180000Qty += 1
			}else if diamond > 180000 && diamond <= 360000{
				diamond180000t360000Qty += 1
			}else{
				diamond360000Qty += 1
			}

		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_orders_range_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, diamond14000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, diamond14000t180000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, diamond180000t360000Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 4, diamond360000Qty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeOrdersRangeDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_orders_range_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		diamond14000Qty, diamond14000t180000Qty, diamond180000t360000Qty, diamond360000Qty)	
	}

	

}
