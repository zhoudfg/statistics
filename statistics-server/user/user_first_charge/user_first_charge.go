package user_first_charge

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dao/viva_vn_api_dao"
	"github.com/statistics/lib/dao/viva_vn_statistics_dao"
	"github.com/statistics/lib/db"
)

// day表示几天前,0为当日
func CountFirstCharge(day int64) {

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
		// 获取用户历史首充日期
		chargeOrders := viva_vn_api_dao.ChargeOrdersDao.GetFirstChargeHistoryDate(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`user_id`, `price`, `price_unit`, `diamond`, `product_id`, `day`) values ", "user_first_charge_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

        totalCount := len(chargeOrders)

		if totalCount == 0 {
			return;
		}
		count := 0
		for _, item := range chargeOrders {
			count += 1
			if totalCount != count{
				buffer.WriteString(fmt.Sprintf(`("%d","%f","%s","%d","%s","%d"),`, item.UserId, item.Price, item.PriceUnit.String, item.Diamond, item.ProductId.String, item.CreateTime))
			}else{
				buffer.WriteString(fmt.Sprintf(`("%d","%f","%s","%d","%s","%d")`, item.UserId, item.Price, item.PriceUnit.String, item.Diamond, item.ProductId.String, item.CreateTime))
			}
		}
		
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE user_id = values(user_id);`))

		viva_vn_statistics_dao.UserFirstChargeLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_first_charge_log day:", date)
	}

}
