package main

import (
	"github.com/robfig/cron/v3"
	"github.com/rroy233/logger"
	"github.com/rroy233/neuDailyReport/reportClient"
	"time"
)

var crontab *cron.Cron

func initCron() {
	crontab = cron.New(cron.WithLocation(time.FixedZone("CST", 8*3600)))
	var err error
	//6点健康上报，7点/12点/19点体温上报
	_, err = crontab.AddFunc("00 6 * * ?", cronAutoDo)
	_, err = crontab.AddFunc("58 7 * * ?", cronAutoDo)
	_, err = crontab.AddFunc("02 12 * * ?", cronAutoDo)
	_, err = crontab.AddFunc("03 19 * * ?", cronAutoDo)
	if err != nil {
		logger.Error.Println("cron初始化失败:", err)
		return
	}
	crontab.Start()
	logger.Info.Println("cron初始化成功")
	return
}

func cronAutoDo() {
	period := reportClient.All
	switch time.Now().Hour() {
	case 6:
		period = reportClient.HealthReport
	case 7:
		period = reportClient.Morning
	case 12:
		period = reportClient.AfterNoon
	case 19:
		period = reportClient.Evening
	}
	for i, _ := range config.StudentList {
		queue <- task{
			accountIndex: i,
			period:       period,
		}
	}
	return
}
