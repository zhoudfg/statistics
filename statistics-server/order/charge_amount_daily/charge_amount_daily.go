package charge_amount

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
		// 获取用户首充时间
		userFirstChargeAmount := viva_vn_statistics_dao.UserFirstOpenDailyLogDao.GetUserFirstChargeUserAmountByTime(db.GetFactory("viva_vn_statistics").GetSession(), startTime, endTime)
		// 获取充值订单
		chargeOrders := viva_vn_api.ChargeOrdersDao.GetChargeOrders2ByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var chargeOrderQty, chargeCompleteOrders, firstChargeUsers int

		var firstChargeAmount float64
		chargeUsers := make(map[int]int, 0)
        chargeCompleteUsers := make(map[int]int, 0)

		for _, qty := range userFirstChargeAmount{
			firstChargeAmount += qty
		}

		firstChargeUsers = len(userFirstChargeAmount)
	
		for _, item := range chargeOrders {
			chargeOrderQty += 1
			chargeUsers[int(item.UserId)] = int(item.UserId)
		    if item.Status == 8 {
				chargeCompleteOrders += 1
				chargeCompleteUsers[int(item.UserId)] = int(item.UserId)
			}
		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into charge_amount_daily_log (`day`, `charge_users`, `charge_complete_users`, `charge_orders`, `charge_complete_orders`, `first_charge_users`, `first_charge_amount`) values ('%s',%d,%d,%d,%d,%d,%f) ON DUPLICATE KEY UPDATE charge_users=%d,charge_complete_users=%d,charge_orders=%d,charge_complete_orders=%d,first_charge_users=%d,first_charge_amount=%f;", 
			 date, len(chargeUsers), len(chargeCompleteUsers), chargeOrderQty, chargeCompleteOrders, firstChargeUsers, firstChargeAmount, len(chargeUsers), len(chargeCompleteUsers), chargeOrderQty, chargeCompleteOrders, firstChargeUsers, firstChargeAmount)

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
			return;
		}

		viva_vn_statistics_dao.ChargeAmountDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_amount_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		len(chargeUsers), len(chargeCompleteUsers), chargeOrderQty, chargeCompleteOrders, firstChargeUsers, firstChargeAmount)	
	}

	

}
