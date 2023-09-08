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
		chargeOrderDiamond := viva_vn_api.ChargeOrdersDao.GetChargeOrdersDiamondByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var diamond1400AndroidQty, diamond7000AndroidQty, diamond14000AndroidQty, diamond100000AndroidQty, diamond180000AndroidQty, diamond360000AndroidQty, totalAndroidQty,
		    diamond1400IosQty, diamond7000IosQty, diamond14000IosQty, diamond100000IosQty, diamond180000IosQty, diamond360000IosQty, totalIosQty int64
	
		for _, item := range chargeOrderDiamond {
			if item.Os.String == "android" {
				if item.Diamond >= 1400 && item.Diamond < 7000 {
					diamond1400AndroidQty += 1
				}else if item.Diamond >= 7000 && item.Diamond < 14000{
					diamond7000AndroidQty += 1
				}else if item.Diamond >= 14000 && item.Diamond < 100000{
					diamond14000AndroidQty += 1
				}else if item.Diamond >= 100000 && item.Diamond < 180000{
					diamond100000AndroidQty += 1
				}else if item.Diamond >= 180000 && item.Diamond < 360000{
					diamond180000AndroidQty += 1
				}else if item.Diamond >= 360000{
					diamond360000AndroidQty += 1
				}
			}else if item.Os.String == "ios" {
				if item.Diamond >= 1400 && item.Diamond < 7000 {
					diamond1400IosQty += 1
				}else if item.Diamond >= 7000 && item.Diamond < 14000{
					diamond7000IosQty += 1
				}else if item.Diamond >= 14000 && item.Diamond < 100000{
					diamond14000IosQty += 1
				}else if item.Diamond >= 100000 && item.Diamond < 180000{
					diamond100000IosQty += 1
				}else if item.Diamond >= 180000 && item.Diamond < 360000{
					diamond180000IosQty += 1
				}else if item.Diamond >= 360000{
					diamond360000IosQty += 1
				}
			}
		}

		totalAndroidQty = diamond1400AndroidQty + diamond7000AndroidQty + diamond14000AndroidQty + diamond100000AndroidQty + diamond180000AndroidQty + diamond360000AndroidQty
		totalIosQty = diamond1400IosQty + diamond7000IosQty + diamond14000IosQty + diamond100000IosQty + diamond180000IosQty + diamond360000IosQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_orders_diamond_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, diamond1400AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, diamond7000AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, diamond14000AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, diamond100000AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, diamond180000AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, diamond360000AndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, totalAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, diamond1400IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, diamond7000IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 10, diamond14000IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 11, diamond100000IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 12, diamond180000IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 13, diamond360000IosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 14, totalIosQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeOrdersDiamondDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_orders_diamond_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		diamond1400AndroidQty, diamond7000AndroidQty, diamond14000AndroidQty, diamond100000AndroidQty, diamond180000AndroidQty, diamond360000AndroidQty, totalAndroidQty,
		diamond1400IosQty, diamond7000IosQty, diamond14000IosQty, diamond100000IosQty, diamond180000IosQty, diamond360000IosQty, totalIosQty)	
	}

	

}
