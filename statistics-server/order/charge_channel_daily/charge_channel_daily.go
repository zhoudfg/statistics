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

		var upayQty, upayDiamondQty, payermaxQty, payermaxDiamondQty, ipasspayQty, ipasspayDiamondQty, googleQty, googleDiamondQty, appleQty, appleDiamondQty, totalDiamond int64
	
		for _, item := range chargeChannels {
			switch item.Channel {
				case "upay":
					upayQty += 1
					upayDiamondQty += item.Diamond
				case "payermax":
					payermaxQty += 1
					payermaxDiamondQty += item.Diamond
				case "ipasspay":
					ipasspayQty += 1
					ipasspayDiamondQty += item.Diamond
				case "google":
					googleQty += 1
					googleDiamondQty += item.Diamond
				case "apple":
					appleQty += 1
					appleDiamondQty += item.Diamond
			}
		}

		totalDiamond = upayDiamondQty + payermaxDiamondQty + ipasspayDiamondQty + googleDiamondQty + appleDiamondQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_channel_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, upayQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, upayDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, payermaxQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, payermaxDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, ipasspayQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, ipasspayDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, googleQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, googleDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, appleQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 10, appleDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 11, totalDiamond))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.ChargeChannelDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_channel_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
				upayQty, upayDiamondQty, payermaxQty, payermaxDiamondQty, ipasspayQty, ipasspayDiamondQty, googleQty, googleDiamondQty,
				appleQty, appleDiamondQty, totalDiamond )	
	}

	

}
