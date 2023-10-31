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

		googleUsdChargeUsers := make(map[int]int, 0)
		googleUsdChargeCompleteUsers := make(map[int]int, 0)
		var googleUsdChargeAmount float64

		momoUsdChargeUsers := make(map[int]int, 0)
		momoUsdChargeCompleteUsers := make(map[int]int, 0)
		var momoUsdChargeAmount float64

		vietnamVTTUsdChargeUsers := make(map[int]int, 0)
		vietnamVTTUsdChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVTTUsdChargeAmount float64

		vietnamVMSUsdChargeUsers := make(map[int]int, 0)
		vietnamVMSUsdChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVMUsdChargeAmount float64

		vietnamVNPUsdChargeUsers := make(map[int]int, 0)
		vietnamVNPUsdChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVNPUsdChargeAmount float64

		googleVndChargeUsers := make(map[int]int, 0)
		googleVndChargeCompleteUsers := make(map[int]int, 0)
		var googleVndChargeAmount float64

		momoVndChargeUsers := make(map[int]int, 0)
		momoVndChargeCompleteUsers := make(map[int]int, 0)
		var momoVndChargeAmount float64

		vietnamVTTVndChargeUsers := make(map[int]int, 0)
		vietnamVTTVndChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVTTVndChargeAmount float64

		vietnamVMSVndChargeUsers := make(map[int]int, 0)
		vietnamVMSVndChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVMVndChargeAmount float64

		vietnamVNPVndChargeUsers := make(map[int]int, 0)
		vietnamVNPVndChargeCompleteUsers := make(map[int]int, 0)
		var vietnamVNPVndChargeAmount float64
	
		for _, item := range chargeOrders {

			if item.PriceUnit.String == "USD" {
				if item.ChargeType.String == "google" {
					googleUsdChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							googleUsdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							googleUsdChargeAmount += item.Price
						}
				}
				switch item.Channel.String {	
					case "momo":
						momoUsdChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							momoUsdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							momoUsdChargeAmount += item.Price
						}
					case "vietnamVTT":
						vietnamVTTUsdChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVTTUsdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVTTUsdChargeAmount += item.Price
						}
					case "vietnamVMS":
						vietnamVMSUsdChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVMSUsdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVMUsdChargeAmount += item.Price
						}
					case "vietnamVNP":
						vietnamVNPUsdChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVNPUsdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVNPUsdChargeAmount += item.Price
						}	
				}
			}else if item.PriceUnit.String == "VND" {
				if item.ChargeType.String == "google" {
					googleVndChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							googleVndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							googleVndChargeAmount += item.Price
						}
				}
				switch item.Channel.String {	
					case "momo":
						momoVndChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							momoVndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							momoVndChargeAmount += item.Price
						}
					case "vietnamVTT":
						vietnamVTTVndChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVTTVndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVTTVndChargeAmount += item.Price
						}
					case "vietnamVMS":
						vietnamVMSVndChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVMSVndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVMVndChargeAmount += item.Price
						}
					case "vietnamVNP":
						vietnamVNPVndChargeUsers[int(item.UserId)] = int(item.UserId)
						if item.Status == 8 {
							vietnamVNPVndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
							vietnamVNPVndChargeAmount += item.Price
						}	
				}
			}	
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `charge_users`, `charge_complete_users`, `charge_amount`, `price_unit`) values ", "charge_channel_amount_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}
	
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 1, len(googleUsdChargeUsers), len(googleUsdChargeCompleteUsers), googleUsdChargeAmount, 1))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 2, len(momoUsdChargeUsers), len(momoUsdChargeCompleteUsers),momoUsdChargeAmount, 1))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 3, len(vietnamVTTUsdChargeUsers), len(vietnamVTTUsdChargeCompleteUsers), vietnamVTTUsdChargeAmount, 1))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 4, len(vietnamVMSUsdChargeUsers), len(vietnamVMSUsdChargeCompleteUsers), vietnamVMUsdChargeAmount, 1))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 5, len(vietnamVNPUsdChargeUsers), len(vietnamVNPUsdChargeCompleteUsers), vietnamVNPUsdChargeAmount, 1))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 1, len(googleVndChargeUsers), len(googleVndChargeCompleteUsers), googleVndChargeAmount, 2))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 2, len(momoVndChargeUsers), len(momoVndChargeCompleteUsers),momoVndChargeAmount, 2))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 3, len(vietnamVTTVndChargeUsers), len(vietnamVTTVndChargeCompleteUsers), vietnamVTTVndChargeAmount, 2))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d"),`, date, 4, len(vietnamVMSVndChargeUsers), len(vietnamVMSVndChargeCompleteUsers), vietnamVMVndChargeAmount, 2))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d","%d","%f","%d")`, date, 5, len(vietnamVNPVndChargeUsers), len(vietnamVNPVndChargeCompleteUsers), vietnamVNPVndChargeAmount, 2))
		
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE charge_users = values(charge_complete_users),charge_complete_users = values(charge_complete_users),charge_amount = values(charge_amount);`))

		viva_vn_statistics_dao.ChargeChannelAmountDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_channel_amount_daily_log day:", date, " startTime:", startTime, " endTime:", endTime)
	}

	

}
