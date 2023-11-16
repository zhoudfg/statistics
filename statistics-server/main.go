package main

import (
	"time"

	"github.com/robfig/cron"
	"github.com/statistics/lib/conf"
	"github.com/statistics/lib/db"
	liblog "github.com/statistics/lib/log"
	"github.com/statistics/lib/mymongoclient"
	"github.com/statistics/statistics-server/diamond/diamond_consume_daily"
	"github.com/statistics/statistics-server/diamond/user_diamond_consume_daily"

	// "github.com/statistics/statistics-server/diamond/diamond_output_daily"
	charge_amount "github.com/statistics/statistics-server/order/charge_amount_daily"
	"github.com/statistics/statistics-server/order/charge_channel_amount_daily"
	"github.com/statistics/statistics-server/order/charge_channel_daily"
	"github.com/statistics/statistics-server/order/charge_orders_daily"
	"github.com/statistics/statistics-server/order/charge_orders_diamond_daily"
	"github.com/statistics/statistics-server/order/charge_orders_range_daily"
	"github.com/statistics/statistics-server/order/vip_orders_daily"
	"github.com/statistics/statistics-server/order/vip_orders_days_daily"
	"github.com/statistics/statistics-server/song/song"
	"github.com/statistics/statistics-server/song/song_interaction"
	"github.com/statistics/statistics-server/song/song_release"
	app_avg_use_time "github.com/statistics/statistics-server/user/app_avg_use_time_daily"
	"github.com/statistics/statistics-server/user/user_activity_age_daily"
	"github.com/statistics/statistics-server/user/user_activity_hourly"
	"github.com/statistics/statistics-server/user/user_first_charge"
	"github.com/statistics/statistics-server/user/user_register_daily"
	"github.com/statistics/statistics-server/user/version_activity_daily"
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

	// 第一次先将所有的数据统计一遍，处理后下次重启注掉这段代码
	var day int64 = 120
	for ; day >= 0; day-- {
		// 记录用户历史首充时间
		user_first_charge.CountFirstCharge(day)
		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(day)
		// 每日登录年龄用户数（区分年龄）
		// user_activity_age_daily.CountDaily(day)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(day)
		// 每日充值订单状态次数
		// charge_orders_daily.CountDaily(day)
		// 每日充值钻石档位
		// charge_orders_diamond_daily.CountDaily(day)
		// 每日充值订单钻石范围用户数
		// charge_orders_range_daily.CountDaily(day)
		// 每日vip订单数
		// vip_orders_daily.CountDaily(day)
		// 每日vip订单天数
		// vip_orders_days_daily.CountDaily(day)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(day)
		// 每日钻石产出类型
		// diamond_output_daily.CountDaily(day)
		// 每日作品统计
		song.CountDaily(day)
		// 每日各版本活动分布
		// version_activity_daily.CountDaily(day)
		// 每日用户充值金额
		// charge_amount.CountDaily(day)
		// 每日用户各渠道充值金额
		// charge_channel_amount_daily.CountDaily(day)
		// APP人均使用时长
		// app_avg_use_time.CountDaily(day)
		// 作品发布统计
		song_release.CountDaily(day)
		// 作品互动数据统计
		song_interaction.CountDaily(day)
		// 钻石消耗类型
		user_diamond_consume_daily.CountDaily(day)

		
		// 避免操作数据库过快休眠1s
		time.Sleep(time.Second)
	}

	// 第一次先将所有的数据统计一遍，处理后下次重启注掉这段代码
	var hour int64 = 1440
	for ; hour >= 0; hour-- {
		// 用户时活统计
		// user_activity_hourly.CountHourly(hour)
	
		// 避免操作数据库过快休眠1s
		// time.Sleep(time.Second)
	}

	//每隔一小时统计一次
	todaySpec := "0 18 * * * ?"
	c.AddFunc(todaySpec, func() {
		// 记录用户历史首充时间
		user_first_charge.CountFirstCharge(0)

		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(0)
		// 每日登录年龄用户数（区分年龄）
		user_activity_age_daily.CountDaily(0)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(0)
		// 每日充值订单状态次数
		charge_orders_daily.CountDaily(0)
		// 每日充值钻石档位
		charge_orders_diamond_daily.CountDaily(0)
		// 每日充值订单钻石范围用户数
		charge_orders_range_daily.CountDaily(0)
		// 每日vip订单数
		vip_orders_daily.CountDaily(0)
		// 每日vip订单天数
		vip_orders_days_daily.CountDaily(0)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(0)
		// 每日钻石产出类型
		// diamond_output_daily.CountDaily(0)
		// 每日作品统计
		song.CountDaily(0)
		// 每日各版本活动分布
		version_activity_daily.CountDaily(0)
		// 每日用户充值金额
		charge_amount.CountDaily(0)
		// 每日用户各渠道充值金额
		charge_channel_amount_daily.CountDaily(0)
		// APP人均使用时长
		app_avg_use_time.CountDaily(0)
		// 作品发布统计
		song_release.CountDaily(0)
		// 作品互动数据统计
		song_interaction.CountDaily(0)
		// 钻石消耗类型
		user_diamond_consume_daily.CountDaily(0)
		
	})

	// 凌晨再统计一次昨天的数据
	yesterdaySpec := "0 30 1 * * ?"
	c.AddFunc(yesterdaySpec, func() {
		// 记录用户历史首充时间
		user_first_charge.CountFirstCharge(1)
		// 用户日活统计
		// user_activity_daily.CountDaily(1)
		// 每日用户注册数（区分性别）
		user_register_daily.CountDaily(1)
		// 每日登录年龄用户数（区分年龄）
		user_activity_age_daily.CountDaily(1)
		// 每日充值渠道次数（区分渠道、钻石）
		charge_channel_daily.CountDaily(1)
		// 每日充值订单状态次数
		charge_orders_daily.CountDaily(1)
		// 每日充值钻石档位
		charge_orders_diamond_daily.CountDaily(1)
		// 每日充值订单钻石范围用户数
		charge_orders_range_daily.CountDaily(1)
		// 每日vip订单数
		vip_orders_daily.CountDaily(1)
		// 每日vip订单天数
		vip_orders_days_daily.CountDaily(1)
		// 每日钻石消耗类型
		diamond_consume_daily.CountDaily(1)
		// 每日钻石产出类型
		// diamond_output_daily.CountDaily(1)
		// 每日作品统计
		song.CountDaily(1)
		// 每日各版本活动分布
		version_activity_daily.CountDaily(1)
		// 每日用户充值金额
		charge_amount.CountDaily(1)
		// 每日用户各渠道充值金额
		charge_channel_amount_daily.CountDaily(1)
		// APP人均使用时长
		app_avg_use_time.CountDaily(1)
		// 作品发布统计
		song_release.CountDaily(1)
		// 作品互动数据统计
		song_interaction.CountDaily(1)
		// 钻石消耗类型
		user_diamond_consume_daily.CountDaily(1)
	})

	// 每隔半小时统计一次
	halfHourSpec := "0 */30 * * * ?"
	c.AddFunc(halfHourSpec, func() {
		// 用户时活统计
		user_activity_hourly.CountHourly(0)
	})

	// 每小时的30分统计上一小时数据
	lastHourSpec := "0 30 * * * ?"
	c.AddFunc(lastHourSpec, func() {
		// 用户时活统计
		user_activity_hourly.CountHourly(1)
	})

	c.Start()
	<-exitProgram
}
