package reportClient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neucn/neugo"
	"github.com/rroy233/neuDailyReport/util"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
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

func New(id, pwd string, pwdEncoded bool) (*reportClient, error) {
	c := new(reportClient)
	c.StuId = id
	if pwdEncoded {
		pwdR, err := base64.StdEncoding.DecodeString(pwd)
		if err != nil {
			return nil, errors.New("base64解码失败:" + err.Error())
		}
		c.Password = string(pwdR)
	} else {
		c.Password = pwd
	}
	c.httpClient = &http.Client{
		Jar: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 10 * time.Second,
	}
	return c, nil
}

func (rc reportClient) writeCookie(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: rc.cookies["XSRF-TOKEN"]})
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: rc.cookies["PHPSESSID"]})
	req.AddCookie(&http.Cookie{Name: "laravel3_session", Value: rc.cookies["laravel3_session"]})
}

func (rc *reportClient) Login() error {
	jar, _ := cookiejar.New(nil)
	casClient := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}
	err := neugo.Use(casClient).WithAuth(rc.StuId, rc.Password).Login(neugo.CAS)
	if err != nil {
		return err
	}
	_, err = casClient.Get("https://pass.neu.edu.cn/tpass/login?service=https%3A%2F%2Fe-report.neu.edu.cn%2Flogin%2Fneupass%2Fcallback")
	if err != nil {
		return err
	}

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
	periodName := ""
	switch period {
	case Morning:
		periodName = "早"
	case AfterNoon:
		periodName = "午"
	case Evening:
		periodName = "晚"
	}

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
	req.Header.Set("Authority", "e-report.neu.edu.cn")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"92\", \" Not A;Brand\";v=\"99\", \"Microsoft Edge\";v=\"92\"")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36 Edg/92.0.902.78")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://e-report.neu.edu.cn")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://e-report.neu.edu.cn/mobile/notes/create")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		log.Printf("[%s][体温上报-%s]上报失败-HTTP请求错误：\n", rc.StuId, periodName)
		return
	}

	if resp.StatusCode == 302 && resp.Header.Get("Location") == "https://e-report.neu.edu.cn/inspection/items" {
		log.Printf("[%s][体温上报-%s]上报成功！\n", rc.StuId, periodName)
	} else {
		log.Printf("[%s][体温上报-%s]上报失败！可能是当前时段无法上报，请稍后再试。\n", rc.StuId, periodName)
	}
}

func (rc *reportClient) QueryStudentInfo() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://e-report.neu.edu.cn/api/profiles/%s", rc.StuId), nil)
	if err != nil {
		panic(err)
	}

	rc.writeCookie(req)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}

	stuInfo := new(StudentInfo)
	err = json.Unmarshal([]byte(util.ReadBody(resp)), stuInfo)
	if err != nil {
		return err
	}
	rc.StudentInfo = stuInfo
	return err
}

func (rc reportClient) ReportHealth() error {
	err := rc.QueryStudentInfo()
	if err != nil {
		return err
	}

	oldCredit := rc.StudentInfo.Data.Credits

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
	form.Travels = make([]interface{}, 0)
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

	req.Header.Set("Authority", "e-report.neu.edu.cn")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"92\", \" Not A;Brand\";v=\"99\", \"Microsoft Edge\";v=\"92\"")
	req.Header.Set("Dnt", "1")
	req.Header.Set("X-Xsrf-Token", rc.cookies["XSRF-TOKEN"])
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36 Edg/92.0.902.78")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://e-report.neu.edu.cn")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://e-report.neu.edu.cn/mobile/notes/create")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5")

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		_ = rc.QueryStudentInfo()
		if oldCredit < rc.StudentInfo.Data.Credits {
			log.Printf("[%s][健康上报]已为%s上报成功，当前积分+10=%d！\n", rc.StuId, rc.StudentInfo.Data.Xingming, rc.StudentInfo.Data.Credits)
		} else {
			log.Printf("[%s][健康上报]%s重复上报，当前积分%d！\n", rc.StuId, rc.StudentInfo.Data.Xingming, rc.StudentInfo.Data.Credits)
		}

	} else {
		log.Printf("[%s][健康上报]为%s上报失败！\n", rc.StuId, rc.StudentInfo.Data.Xingming)
	}
	return err
}
