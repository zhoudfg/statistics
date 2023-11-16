package song_interaction

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

		songCommentLists := viva_vn_api.SongCommentDao.GetSongCommentByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)
		songGiftConsumelists := viva_vn_api.UserGiftConsumeRecordDao.GetSongSendGiftByTime(db.GetFactory("viva_vn_api").GetSession(), startTime, endTime)

		robotUsers := viva_vn_api.UserDao.GetRobotUsers(db.GetFactory("viva_vn_api").GetSession())

		songCommentUsers := make(map[int]int, 0)
		
		songSendGiftUsers := make(map[int]int, 0)
		songSendGiftQty := 0

		songSendPayGiftUsers := make(map[int]int, 0)
		songSendPayGiftDiamondQty := 0

		songSendFreeGiftUsers := make(map[int]int, 0)
		songSendFreeGiftQty := 0
		
		for _, item := range songGiftConsumelists {
			_, ok := robotUsers[item.UserId]
			if ok {
				continue
			}
			songSendGiftUsers[item.UserId] = item.UserId
			songSendGiftQty += item.GiftQty

			if item.GiftPrice > 0 {
				songSendPayGiftUsers[item.UserId] = item.UserId
				songSendPayGiftDiamondQty += item.GiftPrice
			}else{
				songSendFreeGiftUsers[item.UserId] = item.UserId
				songSendFreeGiftQty += item.GiftQty
			}
		}

		for _, item := range songCommentLists {
			_, ok := robotUsers[item.UserId]
			if ok {
				continue
			}
			songCommentUsers[item.UserId] = item.UserId
		}

		var buffer bytes.Buffer

        sql := fmt.Sprintf("insert into `%s` (`day`, `type`, `qty`) values ", "song_interaction_daily_log")

		if _, err := buffer.WriteString(sql); err != nil {
			log.Error(err)
		}

		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 1, len(songCommentUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 2, len(songCommentLists)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 3, len(songSendGiftUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 4, songSendGiftQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 5, len(songSendPayGiftUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 6, songSendPayGiftDiamondQty))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d"),`, date, 7, len(songSendFreeGiftUsers)))
		buffer.WriteString(fmt.Sprintf(`("%s","%d","%d")`, date, 8, songSendFreeGiftQty))
		buffer.WriteString(fmt.Sprintf(`ON DUPLICATE KEY UPDATE qty = values(qty);`))
		
		viva_vn_statistics_dao.SongInteractionDailyLogDao.BatchInsertOrUpdate(db.GetFactory("viva_vn_statistics").GetSession(), buffer.String())

		log.Info("CountDaily song_interaction_daily_log day:", date, " startTime:", startTime, " endTime:", endTime, " result:",
		len(songCommentUsers), len(songCommentLists), len(songSendGiftUsers), songSendGiftQty, len(songSendPayGiftUsers), songSendPayGiftDiamondQty,len(songSendFreeGiftUsers), songSendFreeGiftQty)
	}

	

}
