package db

import (
	"github.com/statistics/lib/dbsession"

	_ "github.com/go-sql-driver/mysql"
)

var Factory *dbsession.DBSessionFactory
var mapFactory map[string]*dbsession.DBSessionFactory

func init() {
	mapFactory = make(map[string]*dbsession.DBSessionFactory, 0)
}

func Init(dbName string, url string) {
	//var err error
	//fmt.Println("mysql=", url)
	factory, err := dbsession.NewDBSessionFactory("mysql", url)
	if err != nil {
		panic(err)
	}

	mapFactory[dbName] = factory
	if dbName == "viva_vn_api" {
		Factory = factory
	}
}

func GetFactory(dbName string) *dbsession.DBSessionFactory {
	return mapFactory[dbName]
}
