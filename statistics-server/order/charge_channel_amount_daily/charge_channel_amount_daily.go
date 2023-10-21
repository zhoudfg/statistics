package charge_channel_amount_daily

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
		chargeOrders := viva_vn_api.ChargeOrdersDao.GetChargeOrders2ByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		googleChargeUsers := make(map[int]int, 0)
		googleChargeCompleteUsers := make(map[int]int, 0)
		var googleChargeAmount float64

		momoChargeUsers := make(map[int]int, 0)
		momoChargeCompleteUsers := make(map[int]int, 0)
		var momoChargeAmount float64

		vietnamVTTChargeUsers := make(map[int]int, 0)
		vietnamVTTChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVTTChargeAmount float64

		vietnamVMSChargeUsers := make(map[int]int, 0)
		vietnamVMSChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVMChargeAmount float64

		vietnamVNPChargeUsers := make(map[int]int, 0)
		vietnamVNPChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVNPChargeAmount float64
	
		for _, item := range chargeOrders {
			switch item.Channel {
				case "google":
					googleChargeUsers[int(item.UserId)] = int(item.UserId)
					if item.Status == 8 {
						googleChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
					}
					googleChargeAmount += item.Price
				case "momo":
					momoChargeUsers[int(item.UserId)] = int(item.UserId)
					if item.Status == 8 {
						momoChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
					}
					momoChargeAmount += item.Price
				case "vietnamVTT":
					vietnamVTTChargeUsers[int(item.UserId)] = int(item.UserId)
					if item.Status == 8 {
						vietnamVTTChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
					}
					vietnamVTTChargeAmount += item.Price
				case "vietnamVMS":
					vietnamVMSChargeUsers[int(item.UserId)] = int(item.UserId)
					if item.Status == 8 {
						vietnamVMSChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
					}
					vietnamVMChargeAmount += item.Price
				case "vietnamVNP":
					vietnamVNPChargeUsers[int(item.UserId)] = int(item.UserId)
					if item.Status == 8 {
						vietnamVNPChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
					}
					vietnamVNPChargeAmount += item.Price
			}
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `charge_users`, `charge_complete_users`, `charge_amount`) values ", "charge_channel_amount_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}
		
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f"),`, date, 1, len(googleChargeUsers), len(googleChargeCompleteUsers), googleChargeAmount))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f"),`, date, 2, len(momoChargeUsers), len(momoChargeCompleteUsers),momoChargeAmount))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f"),`, date, 3, len(vietnamVTTChargeUsers), len(vietnamVTTChargeCompleteUsers), vietnamVTTChargeAmount))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f"),`, date, 4, len(vietnamVMSChargeUsers), len(vietnamVMSChargeCompleteUsers), vietnamVMChargeAmount))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f")`, date, 5, len(vietnamVNPChargeUsers), len(vietnamVNPChargeCompleteUsers), vietnamVNPChargeAmount))
		
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE charge_users = values(charge_complete_users),charge_complete_users = values(charge_complete_users);`))

		viva_vn_statistics_dao.ChargeChannelAmountDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_channel_amount_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
		len(googleChargeUsers), len(googleChargeCompleteUsers), googleChargeAmount,
			len(momoChargeUsers), len(momoChargeCompleteUsers), momoChargeAmount, len(vietnamVTTChargeUsers), len(vietnamVTTChargeCompleteUsers), vietnamVTTChargeAmount,
				len(vietnamVMSChargeUsers), len(vietnamVMSChargeCompleteUsers), vietnamVMChargeAmount, len(vietnamVNPChargeUsers), len(vietnamVNPChargeCompleteUsers), vietnamVNPChargeAmount)
	}

}
