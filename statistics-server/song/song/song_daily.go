package song

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
		songs := viva_vn_api.SongDao.GetSongByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		var  singleAndroidQty, chorusAndroidQty, singleIosQty, chorusIosQty int64
	
		for _, item := range songs {
			switch item.IsChorus {
				case 0:
					if item.Os.String == "android" {
						singleAndroidQty += 1
					}else if item.Os.String == "ios" {
						singleIosQty += 1
					}
				case 1:
                    if item.Os.String == "android" {
						chorusAndroidQty += 1
					}else if item.Os.String == "ios" {
						chorusIosQty += 1
					}
			}
		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "song_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, singleAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, chorusAndroidQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, singleIosQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 4, chorusIosQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.SongDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily song_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:", 
		singleAndroidQty, chorusAndroidQty, singleIosQty, chorusIosQty)
	}

	

}
