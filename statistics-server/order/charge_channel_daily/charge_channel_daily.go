package charge_channel_daily

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
		chargeChannels := viva_vn_api.ChargeOrdersDao.GetChargeChannelByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var upayAndroidQty, upayAndroidDiamondQty, payermaxAndroidQty, payermaxAndroidDiamondQty, ipasspayAndroidQty, ipasspayAndroidDiamondQty, googleAndroidQty, googleAndroidDiamondQty, appleAndroidQty, appleAndroidDiamondQty, totalAndroid, totalAndroidDiamond,
		    upayIosQty, upayIosDiamondQty, payermaxIosQty, payermaxIosDiamondQty, ipasspayIosQty, ipasspayIosDiamondQty, googleIosQty, googleIosDiamondQty, appleIosQty, appleIosDiamondQty, totalIos, totalIosDiamond int64
	
		for _, item := range chargeChannels {
			switch item.Channel {
				case "upay":
					if item.Os == "android" {
						upayAndroidQty += 1
						upayAndroidDiamondQty += item.Diamond
					}else if item.Os == "ios" {
						upayIosQty += 1
						upayIosDiamondQty += item.Diamond
					}
				case "payermax":
					if item.Os == "android" {
						payermaxAndroidQty += 1
						payermaxAndroidDiamondQty += item.Diamond
					}else if item.Os == "ios" {
						payermaxIosQty += 1
						payermaxIosDiamondQty += item.Diamond
					}
				case "ipasspay":
					if item.Os == "android" {
						ipasspayAndroidQty += 1
						ipasspayAndroidDiamondQty += item.Diamond
					}else if item.Os == "ios" {
						ipasspayIosQty += 1
						ipasspayIosDiamondQty += item.Diamond
					}
				case "google":
					if item.Os == "android" {
						googleAndroidQty += 1
						googleAndroidDiamondQty += item.Diamond
					}else if item.Os == "ios" {
						googleIosQty += 1
						googleIosDiamondQty += item.Diamond
					}
				case "apple":
					if item.Os == "android" {
						appleAndroidQty += 1
						appleAndroidDiamondQty += item.Diamond
					}else if item.Os == "ios" {
						appleIosQty += 1
						appleIosDiamondQty += item.Diamond
					}
			}
		}
		totalAndroid = upayAndroidQty + payermaxAndroidQty + ipasspayAndroidQty + googleAndroidQty + appleAndroidQty
		totalAndroidDiamond = upayAndroidDiamondQty + payermaxAndroidDiamondQty + ipasspayAndroidDiamondQty + googleAndroidDiamondQty + appleAndroidDiamondQty
		totalIos = upayIosQty + payermaxIosQty + ipasspayIosQty + googleIosQty + appleIosQty
		totalIosDiamond = upayIosDiamondQty + payermaxIosDiamondQty + ipasspayIosDiamondQty + googleIosDiamondQty + appleIosDiamondQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_channel_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, upayAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, upayAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, payermaxAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, payermaxAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, ipasspayAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, ipasspayAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, googleAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, googleAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, appleAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 10, appleAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 11, totalAndroid))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 12, totalAndroidDiamond))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 13, upayIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 14, upayIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 15, payermaxIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 16, payermaxIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 17, ipasspayIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 18, ipasspayIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 19, googleIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 20, googleIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 21, appleIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 22, appleIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 23, totalIos))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 24, totalIosDiamond))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeChannelDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_channel_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		upayAndroidQty, upayAndroidDiamondQty, payermaxAndroidQty, payermaxAndroidDiamondQty, ipasspayAndroidQty, ipasspayAndroidDiamondQty, googleAndroidQty, googleAndroidDiamondQty, appleAndroidQty, appleAndroidDiamondQty, totalAndroid, totalAndroidDiamond,
		    upayIosQty, upayIosDiamondQty, payermaxIosQty, payermaxIosDiamondQty, ipasspayIosQty, ipasspayIosDiamondQty, googleIosQty, googleIosDiamondQty, appleIosQty, appleIosDiamondQty, totalIos, totalIosDiamond)	
	}

	

}
