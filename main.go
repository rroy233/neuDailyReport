package main

import (
	"encoding/json"
	"github.com/rroy233/neuDailyReport/reportClient"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Config struct {
	TerminateWaitTime int `json:"terminate_wait_time"`
	StudentList       []struct {
		StuId    string `json:"stu_id"`
		Password string `json:"password"`
	} `json:"student_list"`
}

var config *Config

func main() {
	loadConfig()
	for _, s := range config.StudentList {
		rc := reportClient.New(s.StuId, s.Password)
		err := rc.Login()
		if err != nil {
			log.Println("[统一身份验证]登录失败：" + err.Error())
			continue
		}
		rc.ReportTemperature(reportClient.Morning)
		rc.ReportTemperature(reportClient.AfterNoon)
		rc.ReportTemperature(reportClient.Evening)

		err = rc.ReportHealth()
		if err != nil {
			log.Printf("[%s][健康上报]发生错误:%s\n", rc.StuId, err.Error())
			continue
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
