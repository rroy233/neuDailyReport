package main

import (
	"encoding/json"
	"flag"
	"github.com/rroy233/neuDailyReport/reportClient"
	"io/ioutil"
	"log"
	"os"
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

var config *Config

func main() {
	loadConfig()

	a := flag.Bool("a", false, "Enable Morning temperature-report.")
	b := flag.Bool("b", false, "Enable Afternoon temperature-report.")
	c := flag.Bool("c", false, "Enable Evening temperature-report.")
	d := flag.Bool("d", false, "Enable HealthInfo-report.")
	t := flag.Bool("t", false, "Test availability of all student accounts ONLY.")
	flag.Parse()

	//无参数时，默认a,b,c,d全为true
	if !*t && !(*a || *b || *c || *d) {
		*a = true
		*b = true
		*c = true
		*d = true
	}

	for _, s := range config.StudentList {
		rc, err := reportClient.New(s.StuId, s.Password, config.PasswordEncoded)
		if err != nil {
			log.Printf("[%s][初始化失败]%s\n", s.StuId, err.Error())
			continue
		}

		err = rc.Login()
		if err != nil {
			log.Printf("[%s][统一身份验证]登录失败：%s\n", s.StuId, err.Error())
			continue
		}
		log.Printf("[%s][统一身份验证]登录成功!\n", s.StuId)

		if *t {
			err = rc.QueryStudentInfo()
			if err != nil || rc.StudentInfo == nil {
				log.Printf("[%s][测试]获取学生信息失败：%s\n", s.StuId, err.Error())
			} else {
				log.Printf("[%s][测试]学生账号有效:%s(%s)\n", s.StuId, rc.StudentInfo.Data.Xingming, rc.StudentInfo.Data.Xuegonghao)
			}
			continue
		}

		if *a {
			rc.ReportTemperature(reportClient.Morning)
		}
		if *b {
			rc.ReportTemperature(reportClient.AfterNoon)
		}
		if *c {
			rc.ReportTemperature(reportClient.Evening)
		}

		if *d {
			err = rc.ReportHealth()
			if err != nil {
				log.Printf("[%s][健康上报]发生错误:%s\n", rc.StuId, err.Error())
				continue
			}
		}

	}

	defer func() {
		if config.TerminateWaitTime > 0 {
			log.Printf("程序将在%d秒后关闭。。。", config.TerminateWaitTime)
			time.Sleep(time.Duration(config.TerminateWaitTime) * time.Second)
		}
	}()
}

func loadConfig() {
	_, err := os.Stat("./config.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("配置文件不存在，请创建config.json进行配置。")
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
