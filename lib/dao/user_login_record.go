package dao

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/statistics/lib/model"
	"github.com/statistics/lib/mymongoclient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userLoginRecordDao struct{}

var UserLoginRecordDao = new(userLoginRecordDao)

func (*userLoginRecordDao) FindList(appId string, startTime, endTime int64) []*model.UserLoginRecord {
	client := mymongoclient.Instance()
	collection := client.Database("tianhang_im").Collection("user_login_record")
	var ctx = context.Background()

	var list = make([]*model.UserLoginRecord, 0)
	filter := bson.D{
		bson.E{Key: "app_id", Value: appId},
		bson.E{Key: "login_time", Value: bson.M{"$gte": startTime, "$lt": endTime}},
	}
	cur, err := collection.Find(ctx, filter, options.Find())
	if err != nil {
		log.Error("collection.FindOne:", err.Error())
		return list
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		item := &model.UserLoginRecord{}
		_ = cur.Decode(item)
		list = append(list, item)
	}

	return list
}
