package viva_vn_api_dao

import (
	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/dbsession"
)

type chargeRulesDao struct{}

var ChargeRulesDao = new(chargeRulesDao)


func (*chargeRulesDao) GetChargeRulesChannel(session *dbsession.DBSession) map[string]string {
	rows, err := session.Query("SELECT product_id, channel FROM charge_rules")
	if err != nil {
		log.Error("Failed to query the user table: ", err.Error())
		return nil
	}
	defer rows.Close()

	var productId string
	var channel string
	channels := make(map[string]string, 0)
	for rows.Next() {
		err := rows.Scan(&productId, &channel)
		if err != nil {
			log.Error("Failed to scan rows: ", err.Error())
			return nil
		}
		channels[productId] = channel
	}

	return channels
}

