package user_register_daily

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
		// 每天注册的用户性别 总用户数、男、女、未修改性别
		users := viva_vn_api.UserDao.GetRegisterUserSexByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var manUserAndroidQty, womenUserAndroidQty, defaultUserAndroidQty,
			manUserIosQty, womenUserIosQty, defaultUserIosQty int64

		for _, item := range users {

			if item.Os.String == "android" {
				switch item.Sex.Int64 {
				case 1:
					manUserAndroidQty += 1
				case 2:
					womenUserAndroidQty += 1
				case 0:
					defaultUserAndroidQty += 1
				}
			} else if item.Os.String == "ios" {
				switch item.Sex.Int64 {
				case 1:
					manUserIosQty += 1
				case 2:
					womenUserIosQty += 1
				case 0:
					defaultUserIosQty += 1
				}
			}
		}
		
		totalUserAndroidQty := manUserAndroidQty + womenUserAndroidQty + defaultUserAndroidQty
		totalUserIosQty := manUserIosQty + womenUserIosQty + defaultUserIosQty

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "user_register_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, totalUserAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, manUserAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, womenUserAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, defaultUserAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, totalUserIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, manUserIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, womenUserIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 8, defaultUserIosQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.UserRegisterDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_register_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
		totalUserAndroidQty, manUserAndroidQty, womenUserAndroidQty, defaultUserAndroidQty, totalUserIosQty,
		manUserIosQty, womenUserIosQty, defaultUserIosQty )
	}

}
