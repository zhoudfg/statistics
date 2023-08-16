package viva_vn_statistics_dao

import (
	"github.com/statistics/lib/dbsession"
	log "github.com/sirupsen/logrus"
)

type diamondConsumeDailyLogDao struct{}

var DiamondConsumeDailyLogDao = new(diamondConsumeDailyLogDao)

func (*diamondConsumeDailyLogDao) BatchInsertOrUpdate(session *dbsession.DBSession, sql string) (int, error) {
	result, err := session.Exec(sql)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return int(id), nil
}

