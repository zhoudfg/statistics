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

		channels := viva_vn_api.ChargeRulesDao.GetChargeRulesChannel(db.GetFactory("viva_vn_api").GetSession())

		var googleAndroidQty, googleAndroidDiamondQty,
			momoAndroidQty, momoAndroidDiamondQty, momoIosQty, momoIosDiamondQty,
			viettelAndroidQty, viettelAndroidDiamondQty, viettelIosQty, viettelIosDiamondQty,
			mobifoneAndroidQty, mobifoneAndroidDiamondQty, mobifoneIosQty, mobifoneIosDiamondQty,
			vinaphoneAndroidQty, vinaphoneAndroidDiamondQty, vinaphoneIosQty, vinaphoneIosDiamondQty int64

		for _, item := range chargeChannels {

            channel, ok := channels[item.ProductId]

			if !ok {
				continue
			}

			switch channel {
				case "google":
					googleAndroidQty += 1
					googleAndroidDiamondQty += item.Diamond
				case "momo":
					if item.Os.String == "android" {
						momoAndroidQty += 1
						momoAndroidDiamondQty += item.Diamond
					} else if item.Os.String == "ios" {
						momoIosQty += 1
						momoIosDiamondQty += item.Diamond
					}
				case "vietnamVTT":
					if item.Os.String == "android" {
						viettelAndroidQty += 1
						viettelAndroidDiamondQty += item.Diamond
					} else if item.Os.String == "ios" {
						viettelIosQty += 1
						viettelIosDiamondQty += item.Diamond
					}
				case "vietnamVMS":
					if item.Os.String == "android" {
						mobifoneAndroidQty += 1
						mobifoneAndroidDiamondQty += item.Diamond
					} else if item.Os.String == "ios" {
						mobifoneIosQty += 1
						mobifoneIosDiamondQty += item.Diamond
					}
				case "vietnamVNP":
					if item.Os.String == "android" {
						vinaphoneAndroidQty += 1
						vinaphoneAndroidDiamondQty += item.Diamond
					} else if item.Os.String == "ios" {
						vinaphoneIosQty += 1
						vinaphoneIosDiamondQty += item.Diamond
					}
			}
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "charge_channel_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, googleAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, googleAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, momoAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, momoAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, momoIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, momoIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, viettelAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 8, viettelAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 9, viettelIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 10, viettelIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 11, mobifoneAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 12, mobifoneAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 13, mobifoneIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 14, mobifoneIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 15, vinaphoneAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 16, vinaphoneAndroidDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 17, vinaphoneIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 18, vinaphoneIosDiamondQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.ChargeChannelDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_channel_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
		googleAndroidQty, googleAndroidDiamondQty,
		momoAndroidQty, momoAndroidDiamondQty, momoIosQty, momoIosDiamondQty,
		viettelAndroidQty, viettelAndroidDiamondQty, viettelIosQty, viettelIosDiamondQty,
		mobifoneAndroidQty, mobifoneAndroidDiamondQty, mobifoneIosQty, mobifoneIosDiamondQty,
		vinaphoneAndroidQty, vinaphoneAndroidDiamondQty, vinaphoneIosQty, vinaphoneIosDiamondQty)
	}

}
