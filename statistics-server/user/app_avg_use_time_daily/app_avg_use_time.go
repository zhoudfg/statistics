package app_avg_use_time

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/dao"
	"github.com/statistics/lib/dao/viva_vn_api_dao"
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
		// 每天活跃的用户
		activityUser := dao.UserLoginRecordDao.FindListByDaily(conf.GetAppId(), startTime, endTime)

		// 当天注册用户
		registerUser1 := viva_vn_api_dao.UserDao.GetUsersByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)
		// 三天内注册用户
		registerUser2 := viva_vn_api_dao.UserDao.GetUsersThreeRegisterByTime(db.GetFactory("viva_vn_api").GetSession(), startTime-(86400*2), endTime)

		// 所有用户使用时长
		registerUserAppUseTime := make(map[int64]int, 0)
		// 当天注册用户的使用时长
		registerUser1AppUseTime := make(map[int64]int, 0)
		// 三天内注册用户的使用时长
		registerUser2AppUseTime := make(map[int64]int, 0)
		// 大于三天注册用户的使用时长
		registerUser3AppUseTime := make(map[int64]int, 0)

	
		for _, item := range activityUser {
			var useTime int64 = 0
			if item.LoginTime >= startTime && item.LoginTime < endTime {
				if item.LogoutTime < endTime {
					useTime = item.LogoutTime - item.LoginTime
				}else if item.LogoutTime > endTime {
					useTime = endTime - item.LoginTime
				}else if item.LogoutTime == 0 {
					if day != 0  {
						useTime = endTime - item.LoginTime
					}else{
						useTime = now - item.LoginTime
					}
				}
				log.Info("useTime+++++", useTime)
			} else if item.LoginTime < startTime && item.LogoutTime == 0 {
				if day != 0  {
					useTime = endTime - startTime
				}else{
					useTime = now - startTime
				}
				log.Info("useTime*****", useTime)
			} else if item.LoginTime < startTime && item.LogoutTime >= startTime {
				if item.LogoutTime < endTime {
					useTime = item.LogoutTime - startTime
				} else {
					useTime = endTime - startTime
				}
				log.Info("useTime=====", useTime)
			}

			
			
			// 判断是否是当天注册的用户
			_, ok1 := registerUser1[int(item.UserId)]
			// 判断是否是三天注册的用户
			_, ok2 := registerUser2[int(item.UserId)]
			// ... 否则为大于三天注册的用户

			_, ok3 := registerUser1AppUseTime[item.UserId]
			_, ok4 := registerUser2AppUseTime[item.UserId]
			_, ok5 := registerUser3AppUseTime[item.UserId]
			_, ok := registerUserAppUseTime[item.UserId]

			if ok1 {
				if ok3 {
					registerUser1AppUseTime[item.UserId] += int(useTime)
				} else {
					registerUser1AppUseTime[item.UserId] = int(useTime)
				}
			}
			if ok2 {
				if ok4 {
					registerUser2AppUseTime[item.UserId] += int(useTime)
				} else {
					registerUser2AppUseTime[item.UserId] = int(useTime)
				}
			}
			if !ok1 && !ok2 {
				if ok5 {
					registerUser3AppUseTime[item.UserId] += int(useTime)
				} else {
					registerUser3AppUseTime[item.UserId] = int(useTime)
				}
			}
			if ok {
				registerUserAppUseTime[item.UserId] += int(useTime)
			} else {
				registerUserAppUseTime[item.UserId] = int(useTime)
			}

		}

		userTotalTime := 0
		user1TotalTime := 0
		user2TotalTime := 0
		user3TotalTime := 0

		for _, useTime := range registerUser1AppUseTime {
			user1TotalTime += useTime
		}
		for _, useTime := range registerUser2AppUseTime {
			user2TotalTime += useTime
		}
		for _, useTime := range registerUser3AppUseTime {
			user3TotalTime += useTime
		}
		for _, useTime := range registerUserAppUseTime {
			userTotalTime += useTime
		}

		log.Info("userTotalTime====", userTotalTime)
		log.Info("registerUserAppUseTime====", len(registerUserAppUseTime))

		user1AvgUseTime := 0
		user2AvgUseTime := 0
		user3AvgUseTime := 0
		userAvgUseTime := 0

		if len(registerUser1AppUseTime) > 0 {
			user1AvgUseTime = user1TotalTime / len(registerUser1AppUseTime)
		}
		if len(registerUser2AppUseTime) > 0 {
			user2AvgUseTime = user2TotalTime / len(registerUser2AppUseTime)
		}
		if len(registerUser3AppUseTime) > 0 {
			user3AvgUseTime = user3TotalTime / len(registerUser3AppUseTime)
		}
		if len(registerUserAppUseTime) > 0 {
			userAvgUseTime = userTotalTime / len(registerUserAppUseTime)
		}

		var buffer bytes.Buffer

		sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "app_avg_use_time_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, userAvgUseTime))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, user1AvgUseTime))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, user2AvgUseTime))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 4, user3AvgUseTime))

		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))

		viva_vn_statistics_dao.AppAvgUseTimeDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily app_avg_use_time_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
			userAvgUseTime, user1AvgUseTime, user2AvgUseTime, user3AvgUseTime)
	}

}
