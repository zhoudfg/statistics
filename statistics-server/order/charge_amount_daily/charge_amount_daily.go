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
		// 获取用户首充
		userFirstCharges := viva_vn_statistics_dao.UserFirstChargeLogDao.GetUserFirstChargeUserAmountByTime(db.GetFactory("viva_vn_statistics").GetSession(), startTime, endTime)
		// 获取充值订单
		chargeOrders := viva_vn_api.ChargeOrdersDao.GetChargeOrders2ByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var vndChargeOrderQty, vndChargeCompleteOrders, vndFirstChargeUsers int
		var vndFirstChargeAmount float64
		var vndChargeAmount float64
		vndChargeUsers := make(map[int]int, 0)
        vndChargeCompleteUsers := make(map[int]int, 0)
		

		var usdChargeOrderQty, usdChargeCompleteOrders, usdFirstChargeUsers int
		var usdFirstChargeAmount float64
		var usdChargeAmount float64
		usdChargeUsers := make(map[int]int, 0)
        usdChargeCompleteUsers := make(map[int]int, 0)

		for _, item := range userFirstCharges{
			switch item.PriceUnit.String{
			case "VND":
				vndFirstChargeAmount += item.Price
				vndFirstChargeUsers += 1
			case "USD":
				usdFirstChargeAmount += item.Price
				usdFirstChargeUsers += 1
			}
			
		}
	
		for _, item := range chargeOrders {
			
			switch item.PriceUnit.String{
			case "VND":
				vndChargeOrderQty += 1
				vndChargeUsers[int(item.UserId)] = int(item.UserId)
		    	if item.Status == 8 {
					vndChargeAmount += item.Price
					vndChargeCompleteOrders += 1
					vndChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
				}
			case "USD":
				usdChargeOrderQty += 1
				usdChargeUsers[int(item.UserId)] = int(item.UserId)
		    	if item.Status == 8 {
					usdChargeAmount += item.Price
					usdChargeCompleteOrders += 1
					usdChargeCompleteUsers[int(item.UserId)] = int(item.UserId)
				}
			}
		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into charge_amount_daily_log (`day`, `type`, `charge_users`, `charge_complete_users`, `charge_orders`, `charge_complete_orders`, `first_charge_users`, `first_charge_amount`, `charge_amount`) values ('%s', %d, %d, %d, %d, %d, %d, %f, %f) ON DUPLICATE KEY UPDATE charge_users=%d, charge_complete_users=%d, charge_orders=%d, charge_complete_orders=%d, first_charge_users=%d, first_charge_amount=%f, charge_amount=%f;", date, 1, len(usdChargeUsers), len(usdChargeCompleteUsers), usdChargeOrderQty, usdChargeCompleteOrders, usdFirstChargeUsers, usdFirstChargeAmount, usdChargeAmount, len(usdChargeUsers), len(usdChargeCompleteUsers), usdChargeOrderQty, usdChargeCompleteOrders, usdFirstChargeUsers, usdFirstChargeAmount, usdChargeAmount)

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
			return;
		}
		sql2 := fmt.Sprintf("insert into charge_amount_daily_log (`day`, `type`, `charge_users`, `charge_complete_users`, `charge_orders`, `charge_complete_orders`, `first_charge_users`, `first_charge_amount`, `charge_amount`) values ('%s', %d, %d, %d, %d, %d, %d, %f, %f) ON DUPLICATE KEY UPDATE charge_users=%d, charge_complete_users=%d, charge_orders=%d, charge_complete_orders=%d, first_charge_users=%d, first_charge_amount=%f, charge_amount=%f;", date, 2, len(vndChargeUsers), len(vndChargeCompleteUsers), vndChargeOrderQty, vndChargeCompleteOrders, vndFirstChargeUsers, vndFirstChargeAmount, vndChargeAmount, len(vndChargeUsers), len(vndChargeCompleteUsers), vndChargeOrderQty, vndChargeCompleteOrders, vndFirstChargeUsers, vndFirstChargeAmount, vndChargeAmount)

		if _, err := buffer.WriteString(sql2); err != nil {
			log.Error(err)
			return;
		}

		viva_vn_statistics_dao.ChargeAmountDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily charge_amount_daily_log day:", date, " startTime:", startTime, " endTime:", endTime)	
	}

	

}
