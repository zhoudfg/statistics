package user_diamond_consume_daily

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
		userDiamondConsumeRecord := viva_vn_api.UserDiamondRecordDao.GetConsumeDiamondRecordByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		totalConsumeQty := 0
		totalConsumeUsers := make(map[int]int, 0)

		sendGiftConsumeQty := 0
		sendGiftConsumeUsers := make(map[int]int, 0)

		vipConsumeQty := 0
		vipConsumeUsers := make(map[int]int, 0)

		mallConsumeQty := 0
		mallConsumeUsers := make(map[int]int, 0)

		for _, item := range userDiamondConsumeRecord {
			totalConsumeQty += int(item.Qty)
			totalConsumeUsers[item.UserId] = item.UserId
			switch item.Type {
				case 200001:
					sendGiftConsumeQty += int(item.Qty)
					sendGiftConsumeUsers[item.UserId] = item.UserId
				case 200002:
					vipConsumeQty += int(item.Qty)
					vipConsumeUsers[item.UserId] = item.UserId
				case 200004:
					mallConsumeQty += int(item.Qty)
					mallConsumeUsers[item.UserId] = item.UserId	
			}	
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "user_diamond_consume_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, totalConsumeQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, len(totalConsumeUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, sendGiftConsumeQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, len(sendGiftConsumeUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, vipConsumeQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, len(vipConsumeUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, mallConsumeQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 8, len(mallConsumeUsers)))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.UserDiamondConsumeDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_diamond_consume_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
	    totalConsumeQty, len(totalConsumeUsers), sendGiftConsumeQty, len(sendGiftConsumeUsers), vipConsumeQty, len(vipConsumeUsers), mallConsumeQty, len(mallConsumeUsers))
	}

}
