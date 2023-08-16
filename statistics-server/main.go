package main

import (
	"time"

	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/db"
	liblog "github.com/statistics/lib/log"
	"github.com/statistics/lib/mymongoclient"
	"github.com/statistics/statistics-server/diamond/diamond_consume_daily"
	"github.com/statistics/statistics-server/diamond/diamond_output_daily"
	"github.com/statistics/statistics-server/order/charge_channel_daily"
	"github.com/statistics/statistics-server/order/charge_orders_daily"
	"github.com/statistics/statistics-server/order/charge_orders_range_daily"
	"github.com/statistics/statistics-server/order/vip_orders_daily"
	"github.com/statistics/statistics-server/order/vip_orders_days_daily"
	"github.com/statistics/statistics-server/song/song"
	"github.com/statistics/statistics-server/user/user_activity_age_daily"
	"github.com/statistics/statistics-server/user/user_activity_hourly"
	"github.com/statistics/statistics-server/user/user_register_daily"

	"github.com/robfig/cron"
)

var (
	exitProgram = make(chan bool)
)

func init() {
	var serverName string = "statistics-server"
	path := serverName

	conf.Init(path)

	liblog.InitLog(conf.GetLog("basedir"), "statistics-server.log", conf.GetLogLevel())
	db.Init("viva_vn_api", conf.GetMysqlSource("viva_vn_api"))
	db.Init("viva_vn_statistics", conf.GetMysqlSource("viva_vn_statistics"))
	mymongoclient.Init("tianhang_im", conf.GetMongo("tianhang_im"))
}

func main() {

	c := cron.New()

	// var month int = 0

	// for ; month >= -25; month-- {
	// 	// vip用户
	// 	vip.CountMonth(month)
	// 	// 避免操作数据库过快休眠1s
	// 	time.Sleep(time.Second)
	// }

	// 第一次先将所有的数据统计一遍，处理后下次重启注掉这段代码
	var day int64 = 4

	for ; day >= 0; day-- {
		// 用户日活统计
		// user_activity_daily.CountDaily(day)
		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(day)
		// 每日登录年龄用户数（区分年龄）
		user_activity_age_daily.CountDaily(day)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(day)
		// 每日充值订单状态次数（区分状态）
		charge_orders_daily.CountDaily(day)
		// 每日充值钻石档位
		charge_orders_daily.CountDaily(day)
		// 每日充值订单钻石范围用户数
		charge_orders_range_daily.CountDaily(day)
		// 每日vip订单数
		vip_orders_daily.CountDaily(day)
		// 每日vip订单天数
		vip_orders_days_daily.CountDaily(day)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(day)
		// 每日钻石产出类型
		diamond_output_daily.CountDaily(day)
		// 每日作品统计
		song.CountDaily(day)
		



		// 避免操作数据库过快休眠1s
		time.Sleep(time.Second)
	}

	// 第一次先将所有的数据统计一遍，处理后下次重启注掉这段代码
	var hour int64 = 12

	for ; hour >= 0; hour-- {
		// 用户时活统计
		user_activity_hourly.CountHourly(hour)
	
		// 避免操作数据库过快休眠1s
		time.Sleep(time.Second)
	}

	// 每天早上6点统计一次
	dailySpec := "0 0 6 * * ?"
	c.AddFunc(dailySpec, func() {

	})

	//每隔一小时统计一次
	todaySpec := "0 18 * * * ?"
	c.AddFunc(todaySpec, func() {
		// 用户日活统计
		// user_activity_daily.CountDaily(0)
		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(0)
		// 每日登录年龄用户数（区分年龄）
		user_activity_age_daily.CountDaily(0)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(0)
		// 每日充值订单状态次数（区分状态）
		charge_orders_daily.CountDaily(0)
		// 每日充值钻石档位
		charge_orders_daily.CountDaily(0)
		// 每日充值订单钻石范围用户数
		charge_orders_range_daily.CountDaily(0)
		// 每日vip订单数
		vip_orders_daily.CountDaily(0)
		// 每日vip订单天数
		vip_orders_days_daily.CountDaily(0)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(0)
		// 每日钻石产出类型
		diamond_output_daily.CountDaily(0)
		// 每日作品统计
		song.CountDaily(0)
	})

	// 凌晨再统计一次昨天的数据
	yesterdaySpec := "0 30 1 * * ?"
	c.AddFunc(yesterdaySpec, func() {
		// 用户日活统计
		// user_activity_daily.CountDaily(1)
		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(1)
		// 每日登录年龄用户数（区分年龄）
		user_activity_age_daily.CountDaily(1)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(1)
		// 每日充值订单状态次数（区分状态）
		charge_orders_daily.CountDaily(1)
		// 每日充值钻石档位
		charge_orders_daily.CountDaily(1)
		// 每日充值订单钻石范围用户数
		charge_orders_range_daily.CountDaily(1)
		// 每日vip订单数
		vip_orders_daily.CountDaily(1)
		// 每日vip订单天数
		vip_orders_days_daily.CountDaily(1)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(1)
		// 每日钻石产出类型
		diamond_output_daily.CountDaily(1)
		// 每日作品统计
		song.CountDaily(1)
	})

	// 每月1号再统计一次上月的数据
	lastMonthSpec := "0 30 1 1 * ?"
	c.AddFunc(lastMonthSpec, func() {

	})

	c.Start()
	<-exitProgram
}
