package user_first_open

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

		sql := fmt.Sprintf("insert into `%s` (`user_id`, `type`, `day`, `qty`) values ", "user_first_open_daily_log")

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
				buffer.WriteString(fmt.Sprintf(`("%d","%d","%d","%f"),`, item.UserId, 1, item.CreateTime, item.Price))
			}else{
				buffer.WriteString(fmt.Sprintf(`("%d","%d","%d","%f")`, item.UserId, 1, item.CreateTime, item.Price))
			}
		}
		
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE day = values(day), qty = values(qty);`))

		viva_vn_statistics_dao.UserFirstOpenDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_first_open_daily_log day:", " result:", totalCount)
	}

}
