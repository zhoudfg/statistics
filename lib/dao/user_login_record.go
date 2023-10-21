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

// 每天登录用户(不去重)
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

// 每天登录用户(去重)
func (*userLoginRecordDao) FindListByDistinctUserId(appId string, startTime, endTime int64) map[int64]int64 {
	client := mymongoclient.Instance()
	collection := client.Database("tianhang_im").Collection("user_login_record")
	var ctx = context.Background()
	filter := bson.D{
		bson.E{Key: "app_id", Value: appId},
		bson.E{Key: "login_time", Value: bson.M{"$gte": startTime, "$lt": endTime}},
	}
	userIds, err := collection.Distinct(ctx, "user_id", filter)
	if err != nil {
		log.Error("collection.FindOne:", err.Error())
		return nil
	}

	userIdsMap := make(map[int64]int64)
	for _, userId := range userIds {
		if userId, ok := userId.(int64); ok {
			// 使用默认值示例，您可以根据需求进行调整
			userIdsMap[userId] = userId
		}
	}

	return userIdsMap
}

// 每小时活跃用户
func (*userLoginRecordDao) FindListByHourly(appId string, startTime, endTime int64) []*model.UserLoginRecord {
	client := mymongoclient.Instance()
	collection := client.Database("tianhang_im").Collection("user_login_record")
	var ctx = context.Background()
	var list = make([]*model.UserLoginRecord, 0)
	filter := bson.D{
		bson.E{Key: "app_id", Value: appId},
		bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$gte": startTime, "$lt": endTime}},
				},
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$lt": endTime}},
					bson.E{Key: "logout_time", Value: 0},
				},
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$lt": startTime}},
					bson.E{Key: "logout_time", Value: bson.M{"$gte": startTime}},
				},
			
			},
		},	
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

// 每天活跃用户
func (*userLoginRecordDao) FindListByDaily(appId string, startTime, endTime int64) []*model.UserLoginRecord {
	client := mymongoclient.Instance()
	collection := client.Database("tianhang_im").Collection("user_login_record")
	var ctx = context.Background()
	var list = make([]*model.UserLoginRecord, 0)
	filter := bson.D{
		bson.E{Key: "app_id", Value: appId},
		bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$gte": startTime, "$lt": endTime}},
				},
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$lt": endTime}},
					bson.E{Key: "logout_time", Value: 0},
				},
				bson.D{
					bson.E{Key: "login_time", Value: bson.M{"$lt": startTime}},
					bson.E{Key: "logout_time", Value: bson.M{"$gte": startTime}},
				},
			},
		},	
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


