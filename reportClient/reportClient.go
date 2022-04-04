package reportClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/neucn/neugo"
	"github.com/rroy233/neuDailyReport/util"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	Morning   = 1
	AfterNoon = 2
	Evening   = 3
)

type reportClient struct {
	StuId         string
	Password      string
	httpClient    *http.Client
	formCsrfToken string
	cookies       map[string]string
	StudentInfo   *StudentInfo
}

func New(id, pwd string) *reportClient {
	c := new(reportClient)
	c.StuId = id
	c.Password = pwd
	c.httpClient = &http.Client{
		Jar: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return c
}

func (rc reportClient) writeCookie(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: rc.cookies["XSRF-TOKEN"]})
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: rc.cookies["PHPSESSID"]})
	req.AddCookie(&http.Cookie{Name: "laravel3_session", Value: rc.cookies["laravel3_session"]})
}

func (rc *reportClient) Login() error {
	casClient := neugo.NewSession()
	err := neugo.Use(casClient).WithAuth(rc.StuId, rc.Password).Login(neugo.CAS)
	if err != nil {
		return err
	}
	_, err = casClient.Get("https://pass.neu.edu.cn/tpass/login?service=https%3A%2F%2Fe-report.neu.edu.cn%2Flogin%2Fneupass%2Fcallback")
	if err != nil {
		return err
	}
	log.Println("[统一身份验证]登录成功！")

	cookies := make(map[string]string)
	resp, err := casClient.Get("https://e-report.neu.edu.cn/inspection/items/1/records/create")
	if err != nil {
		panic(err)
	}
	cookies["XSRF-TOKEN"], _ = util.MatchSingle(regexp.MustCompile(`XSRF-TOKEN=(.+?);`), util.JsonEncode(resp.Header.Values("set-cookie")))
	cookies["laravel3_session"], _ = util.MatchSingle(regexp.MustCompile(`laravel3_session=(.+?);`), util.JsonEncode(resp.Header.Values("set-cookie")))
	for _, cookie := range casClient.Jar.Cookies(util.ParseUrl("https://e-report.neu.edu.cn/")) {
		if cookie.Name == "PHPSESSID" {
			cookies["PHPSESSID"] = cookie.Value
		}
	}
	token, _ := util.MatchSingle(regexp.MustCompile(`name="_token" value="(.+?)"`), util.ReadBody(resp))

	rc.cookies = cookies
	rc.formCsrfToken = token
	return nil
}

func (rc reportClient) ReportTemperature(period int) {
	params := url.Values{}
	params.Add("_token", rc.formCsrfToken)
	params.Add("temperature", `36.5`)
	params.Add("suspicious_respiratory_symptoms", `0`)
	params.Add("symptom_descriptions", ``)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://e-report.neu.edu.cn/inspection/items/%d/records", period), body)
	if err != nil {
		panic(err)
	}

	rc.writeCookie(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rc.httpClient.Do(req)

	periodName := ""
	switch period {
	case Morning:
		periodName = "早"
	case AfterNoon:
		periodName = "午"
	case Evening:
		periodName = "晚"
	}
	if resp.StatusCode == 302 {
		log.Printf("[%s][体温上报-%s]上报成功！\n", rc.StuId, periodName)
	} else {
		log.Printf("[%s][体温上报-%s]上报失败！\n", rc.StuId, periodName)
	}
}

func (rc *reportClient) queryStudentInfo() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://e-report.neu.edu.cn/api/profiles/%s", rc.StuId), nil)
	if err != nil {
		panic(err)
	}

	rc.writeCookie(req)

	resp, err := rc.httpClient.Do(req)

	stuInfo := new(StudentInfo)
	err = json.Unmarshal([]byte(util.ReadBody(resp)), stuInfo)
	if err != nil {
		return err
	}
	rc.StudentInfo = stuInfo
	return err
}

func (rc reportClient) ReportHealth() error {
	err := rc.queryStudentInfo()
	if err != nil {
		return err
	}
	form := new(HealthReportForm)
	form.Token = rc.formCsrfToken
	form.JibenxinxiShifoubenrenshangbao = "1"
	form.Profile.Xuegonghao = rc.StuId
	form.Profile.Xingming = rc.StudentInfo.Data.Xingming
	form.Profile.Suoshubanji = rc.StudentInfo.Data.Suoshubanji
	form.JiankangxinxiMuqianshentizhuangkuang = "正常"
	form.XingchengxinxiWeizhishifouyoubianhua = "0"
	form.CrossCity = "无"
	form.Credits = "10"
	form.CityCode = "210000"
	form.ProvinceCode = "210000"
	payloadBytes, err := json.Marshal(form)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", "https://e-report.neu.edu.cn/api/notes", body)
	if err != nil {
		return err
	}
	rc.writeCookie(req)
	resp, err := rc.httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode == 302 {
		_ = rc.queryStudentInfo()
		log.Printf("[%s][健康上报]已为%s上报成功，当前积分为%d！\n", rc.StuId, rc.StudentInfo.Data.Xingming, rc.StudentInfo.Data.Credits)
	} else {
		log.Printf("[%s][健康上报]为%s上报失败！\n", rc.StuId, rc.StudentInfo.Data.Xingming)
	}
	return err
}
