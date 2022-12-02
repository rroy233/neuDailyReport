package main

import (
	"encoding/json"
	"flag"
	"github.com/rroy233/logger"
	"github.com/rroy233/neuDailyReport/reportClient"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Config struct {
	TerminateWaitTime int  `json:"terminate_wait_time"`
	PasswordEncoded   bool `json:"password_encoded"`
	StudentList       []struct {
		StuId    string `json:"stu_id"`
		Password string `json:"password"`
	} `json:"student_list"`
}

type task struct {
	accountIndex int
	period       reportClient.Period
	useWg        bool
	retryCount   int
}

const MaxRetryNum = 10

var config *Config
var queue chan task
var waitGroup sync.WaitGroup

var (
	a *bool
	b *bool
	c *bool
	d *bool
	t *bool
)

func main() {
	loadConfig()

	a = flag.Bool("a", false, "Enable Morning temperature-report.")
	b = flag.Bool("b", false, "Enable Afternoon temperature-report.")
	c = flag.Bool("c", false, "Enable Evening temperature-report.")
	d = flag.Bool("d", false, "Enable HealthInfo-report.")
	t = flag.Bool("t", false, "Test availability of all student accounts ONLY.")
	flag.Parse()

	//log
	logger.New(&logger.Config{
		StdOutput:      true,
		StoreLocalFile: true,
	})

	if *t == true {
		testAccount()
		return
	}

	//启动worker
	go reportWorker()

	//初始化队列
	queue = make(chan task, 10)

	//无参数时，默认常驻运行
	if !*t && !(*a || *b || *c || *d) {
		initCron()
	} else {
		waitGroup = sync.WaitGroup{}
		for i, _ := range config.StudentList {
			waitGroup.Add(1)
			queue <- task{accountIndex: i, period: reportClient.All, useWg: true}
		}
		waitGroup.Wait()
		return
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	<-sigCh

	defer func() {
		if config.TerminateWaitTime > 0 {
			logger.Info.Printf("程序将在%d秒后关闭。。。", config.TerminateWaitTime)
			time.Sleep(time.Duration(config.TerminateWaitTime) * time.Second)
		}
	}()
}

func testAccount() {
	for _, s := range config.StudentList {
		rc, err := reportClient.New(s.StuId, s.Password, config.PasswordEncoded)
		if err != nil {
			logger.Error.Printf("[%s][初始化失败]%s\n", s.StuId, err.Error())
			continue
		}

		err = rc.Login()
		if err != nil {
			logger.Error.Printf("[%s][统一身份验证]登录失败：%s\n", s.StuId, err.Error())
			continue
		}
		logger.Info.Printf("[%s][统一身份验证]登录成功!\n", s.StuId)

		if *t {
			err = rc.QueryStudentInfo()
			if err != nil || rc.StudentInfo == nil {
				logger.Error.Printf("[%s][测试]获取学生信息失败：%s\n", s.StuId, err.Error())
			} else {
				logger.Info.Printf("[%s][测试]学生账号有效:%s(%s)\n", s.StuId, rc.StudentInfo.Data.Xingming, rc.StudentInfo.Data.Xuegonghao)
			}
			continue
		}
	}
}

func reportWorker() {
	for {
		reportTask := <-queue
		s := config.StudentList[reportTask.accountIndex]

		//判断是否超过最大重试次数
		if reportTask.retryCount > MaxRetryNum {
			logger.Error.Printf("[%s][任务已放弃]超过重试次数！！\n", s.StuId)
			continue
		}

		rc, err := reportClient.New(s.StuId, s.Password, config.PasswordEncoded)
		if err != nil {
			logger.Error.Printf("[%s][初始化失败]%s\n", s.StuId, err.Error())
			makeRetry(reportTask)
			continue
		}

		err = rc.Login()
		if err != nil {
			logger.Error.Printf("[%s][统一身份验证]登录失败：%s\n", s.StuId, err.Error())
			makeRetry(reportTask)
			continue
		}
		logger.Info.Printf("[%s][统一身份验证]登录成功!\n", s.StuId)

		//体温上报
		if *a || reportTask.period == reportClient.Morning {
			err = rc.ReportTemperature(reportClient.Morning)
		}
		if *b || reportTask.period == reportClient.AfterNoon {
			err = rc.ReportTemperature(reportClient.AfterNoon)
		}
		if *c || reportTask.period == reportClient.Evening {
			err = rc.ReportTemperature(reportClient.Evening)
		}
		if err != nil {
			logger.Error.Printf(err.Error())
			makeRetry(reportTask)
			continue
		}

		//健康上报
		if *d || reportTask.period == reportClient.HealthReport {
			err = rc.ReportHealth()
			if err != nil {
				logger.Error.Printf("[%s][健康上报]发生错误:%s\n", rc.StuId, err.Error())
				makeRetry(reportTask)
				continue
			}
		}
		if reportTask.useWg == true {
			waitGroup.Done()
		}
	}
}

func makeRetry(reportTask task) {
	reportTask.retryCount++
	queue <- reportTask
	time.Sleep(3 * time.Second)
	return
}

func loadConfig() {
	_, err := os.Stat("./config.json")
	if err != nil {
		if os.IsNotExist(err) {
			logger.FATAL.Fatalf("配置文件不存在，请创建config.json进行配置。")
		}
	}
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	config = new(Config)
	err = json.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
}
