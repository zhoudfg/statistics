package user_first_charge

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dao/viva_vn_api_dao"
	"github.com/statistics/lib/dao/viva_vn_statistics_dao"
	"github.com/statistics/lib/db"
)

// day表示几天前,0为当日
func CountFirstCharge() {
	
	{
		// 获取用户历史首充日期
		userIdDate := viva_vn_api_dao.ChargeOrdersDao.GetFirstChargeHistoryDate(db.GetFactory("viva_vn_api").GetSession())

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`user_id`, `price`, `price_unit`, `diamond`, `product_id`, `day`) values ", "user_first_charge_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

        totalCount := len(userIdDate)

		if totalCount == 0 {
			return;
		}
		count := 0
		for _, item := range userIdDate {
			count += 1
			if totalCount != count{
				buffer.WriteString(fmt.Sprintf(`("%d","%f","%s","%d","%s","%d"),`, item.UserId, item.Price, item.PriceUnit.String, item.Diamond, item.ProductId.String, item.CreateTime))
			}else{
				buffer.WriteString(fmt.Sprintf(`("%d","%f","%s","%d","%s","%d")`, item.UserId, item.Price, item.PriceUnit.String, item.Diamond, item.ProductId.String, item.CreateTime))
			}
		}
		
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE day = values(day), price = values(price), price_unit = values(price_unit), diamond = values(diamond), product_id = values(product_id);`))

		viva_vn_statistics_dao.UserFirstChargeLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_first_charge_log day:", " result:", totalCount)
	}

}
