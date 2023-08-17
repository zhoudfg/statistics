package user

import (
	"bytes"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"
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
		// 当天活跃的用户
		activityUser := dao.UserLoginRecordDao.FindList("10029", startTime, endTime)
		// 当天注册的新用户
		user := viva_vn_api.UserDao.GetNewUserByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var set = mapset.NewSetFromSlice(user)
		var androidNewUserQty, androidOldUserQty, iosNewUserQty, iosOldUserQty int64
		var userIdTeamMap = make(map[int64]int64)
	
		for _, item := range activityUser {
			_, ok := userIdTeamMap[item.UserId]
			if ok {
				continue
			}else{
				userIdTeamMap[item.UserId] = item.UserId
			}
			switch item.Os {
				case "android":
					if set.Contains(item.UserId) {
						androidNewUserQty += 1
					}else{
						androidOldUserQty += 1
					}
				case "ios":
					if set.Contains(item.UserId) {
						iosNewUserQty += 1
					}else{
						iosOldUserQty += 1
					}
			}
		}
		var androidTotalUserQty = androidNewUserQty + androidOldUserQty
		var iosTotalUserQty = iosNewUserQty + iosOldUserQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "user_activity_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, androidNewUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, androidOldUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, androidTotalUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, iosNewUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, iosOldUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 6, iosTotalUserQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.UserActivityDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily user_activity_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
				androidNewUserQty, androidOldUserQty, androidTotalUserQty, iosNewUserQty, iosOldUserQty, iosTotalUserQty)	
	}

	

}
