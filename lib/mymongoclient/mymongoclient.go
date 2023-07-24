package mymongoclient

import (
	"context"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var once sync.Once
var (
	instance       *mongo.Client
	mapMongoClient map[string]*mongo.Client
)

func init() {
	mapMongoClient = make(map[string]*mongo.Client, 0)
}

func Init(dbName string, uri string) {
	var err error
	var client *mongo.Client

	opts := options.Client()
	opts.ApplyURI(uri)
	opts.SetMaxPoolSize(20)
	if client, err = mongo.Connect(context.Background(), opts); err != nil {
		log.Error("mongo connect", err.Error())
		return
	}
	client.Ping(context.Background(), nil)
	mapMongoClient[dbName] = client
}

func GetMongoClientByDB(name string) *mongo.Client {
	return mapMongoClient[name]
}

func Instance() *mongo.Client {
	if instance == nil {
		instance = getMongoClient()
	}
	return instance
}

func getMongoClient() *mongo.Client {
	var err error
	var client *mongo.Client
	uri := conf.GetHost("mongo")
	//uri := "mongodb://127.0.0.1:27017/replicaSet=rs"
	//uri := "mongodb://admin:123456@127.0.0.1:27017/mysm"
	//uri := globalconf.GetHost("mongodb")
	//uri := "mongodb://db0.example.com:27017,db1.example.com:27017,db2.example.com:27017/admin?replicaSet=myRepl"
	//uri := "mongodb://localhost/argos?replicaSet=replset"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}
	opts := options.Client()
	opts.ApplyURI(uri)
	opts.SetMaxPoolSize(20)
	if client, err = mongo.Connect(context.Background(), opts); err != nil {
		log.Error("mongo connect", err.Error())
		return nil
	}
	client.Ping(context.Background(), nil)
	return client
}
