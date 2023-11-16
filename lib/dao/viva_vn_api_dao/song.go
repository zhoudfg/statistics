package viva_vn_api_dao

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type songDao struct{}

var SongDao = new(songDao)

type song struct{
	IsChorus int32
	Os       sql.NullString
	UserId   int
	
}

func (*songDao) GetSongByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*song {
	rows, err := session.Query("SELECT is_chorus, user_register_info.mobile_os AS os, song.user_id as user_id from song LEFT JOIN user_register_info ON user_register_info.user_id = song.user_id where create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the song table: ", err.Error())
		return nil
	}
	defer rows.Close()

	songs := make([]*song, 0)
	for rows.Next() {
		s := &song{}
		err := rows.Scan(&s.IsChorus, &s.Os, &s.UserId)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		songs = append(songs, s)
	}

	return songs
}

func (*songDao) GetSong2ByTime(session *dbsession.DBSession, startTime int64, endTime int64) []*song {
	rows, err := session.Query("SELECT user_id, is_chorus from song where create_time >= ? and create_time < ?", startTime, endTime)
	if err != nil {
		log.Error("Failed to query the song table: ", err.Error())
		return nil
	}
	defer rows.Close()

	songs := make([]*song, 0)
	for rows.Next() {
		s := &song{}
		err := rows.Scan(&s.UserId, &s.IsChorus)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		songs = append(songs, s)
	}

	return songs
}

