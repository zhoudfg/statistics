package viva_vn_api_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type songDao struct{}

var SongDao = new(songDao)

type song struct{
	IsChorus int32
}

func (*songDao) GetSongByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*song {
	rows, err := session.Query("SELECT is_chorus from song where create_time >= ? and create_time < ?", startTime, endTime)

	var songs = make([]*song, 0)
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	for rows.Next() {
		s := &song{}
		err := rows.Scan(&s.IsChorus)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		songs = append(songs, s)
	}

	return songs
}
