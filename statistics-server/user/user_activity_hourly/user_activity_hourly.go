package user_activity_hourly

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/dao"
	"github.com/statistics/lib/dao/viva_vn_statistics_dao"
	"github.com/statistics/lib/db"
)

// day表示几天前,0为当日
func CountHourly(hour int64) {
	timeStr := time.Now().In(time.Local).Format("2006-01-02 15:00:00")
	t, _ := time.Parse("2006-01-02 15:00:00", timeStr)
	startTime := t.Unix() - (8 * 3600) - 3600*hour
	endTime := t.Unix() - (8 * 3600) - (3600 * (hour - 1))
	timeObj := time.Unix(int64(startTime + (8 * 3600)), 0)
	datetime := timeObj.UTC().Format("2006-01-02 15:04:05")

	now := time.Now().Unix()
	if now < endTime {
		endTime = now - 5
	}

	{
		// 每小时活跃的用户
		activityUser := dao.UserLoginRecordDao.FindListByHourly(conf.GetAppId(), startTime, endTime)
		fmt.Println(activityUser)
		
		var androidUserQty, iosUserQty, totalUserQty int64
		var userIdTeamMap = make(map[int64]int64)
	
		for _, item := range activityUser {
			_, ok := userIdTeamMap[item.UserId]
			if ok {
				continue
			}else{
				userIdTeamMap[item.UserId] = item.UserId
			}
			switch item.Os{
				case "android":
					androidUserQty += 1
				case "ios":
					iosUserQty += 1
			}
			
		}

		totalUserQty = androidUserQty + iosUserQty

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`hour`, `type`, `qty`) values ", "user_activity_hourly_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, datetime, 1, androidUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, datetime, 2, iosUserQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, datetime, 3, totalUserQty))

		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.UserActivityHourlyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountHourly user_activity_hourly_log day:", datetime, " startTime:", startTime, " endTime:", endTime, " result:", 
		androidUserQty, iosUserQty, totalUserQty)	
	}

	

}
