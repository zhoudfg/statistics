package user_activity_age_daily

import (
	"bytes"
	"fmt"
	
	// "strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dao"
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
	timeObj := time.Unix(int64(startTime), 0)
	date := timeObj.Format("2006-01-02")

	now := time.Now().Unix()
	if now < endTime {
		endTime = now - 5
	}

	{
		// 当天登录的用户
		activityUserIds := dao.UserLoginRecordDao.FindListByDistinctUserId("10029", startTime, endTime)
		
		var userIds []int64
		for _, userId := range activityUserIds {
			userIds = append(userIds, userId)
		}

		// 获取用户年龄
		userAges := viva_vn_api.UserPersonInfoDao.GetActivityUserAgeByUserIds(db.GetFactory("viva_vn_api").GetSession(), userIds)


		var age18Qty, age18and30Qty, age30and50Qty, age50Qty int64

		for _, userAge := range userAges{
			birthday := int64(userAge.Birthday)
			t := time.Unix(birthday, 0)
			currentDate := time.Now().In(time.Local)
			age := currentDate.Year() - t.Year()
			// 检查是否已过生日，如果生日尚未到达，年龄减少一岁
			if currentDate.YearDay() < t.YearDay() {
				age--
			}
			if age <= 18 {
				age18Qty += 1
			}else if age > 18 && age <= 30{
				age18and30Qty += 1
			}else if age > 30 && age <= 50{
				age30and50Qty += 1
			}else if age > 50{
				age50Qty += 1
			}
		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "user_activity_age_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, age18Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, age18and30Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, age30and50Qty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 4, age50Qty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.UserActivityAgeDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_activity_age_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
				age18Qty, age18and30Qty, age30and50Qty, age50Qty)	
	}

	

}
