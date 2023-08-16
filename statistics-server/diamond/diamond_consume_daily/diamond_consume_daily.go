package diamond_consume_daily

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
	timeStr := time.Now().Format("2006-01-02")
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
		
		userDiamondConsume := viva_vn_api.UserDiamondRecordDao.GetConsumeDiamondTypeQtyByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var sendGiftQty, totalQty int64
	
		for _, item := range userDiamondConsume {
			switch item.Type {
				case 200001:
                    sendGiftQty += item.Qty
			}
		}

		totalQty = sendGiftQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "diamond_consume_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, totalQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 2, sendGiftQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.DiamondConsumeDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily diamond_consume_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		sendGiftQty, totalQty)
	}

	

}
