package viva_vn_api_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type songCommentDao struct{}

var SongCommentDao = new(songCommentDao)

type songComment struct{
	UserId   int
    SongId   int
}

func (*songCommentDao) GetSongCommentByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*songComment {
	rows, err := session.Query("SELECT user_id, song_id from song_comment where create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the song_comment table: ", err.Error())
		return nil
	}
	defer rows.Close()

	songComments := make([]*songComment, 0)
	for rows.Next() {
		s := &songComment{}
		err := rows.Scan(&s.UserId, &s.SongId)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		songComments = append(songComments, s)
	}

	return songComments
}

