package util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func ReadBody(resp *http.Response) (body string) {
	res, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return string(res)
}

func MatchSingle(re *regexp.Regexp, content string) (string, error) {
	matched := re.FindAllStringSubmatch(content, -1)
	if len(matched) < 1 {
		return "", errors.New("errorNoMatched")
	}
	return matched[0][1], nil
}

func ParseUrl(in string) *url.URL {
	u, err := url.Parse(in)
	if err != nil {
		panic(err)
	}
	return u
}

func JsonEncode(data interface{}) string {
	out, _ := json.Marshal(data)
	return string(out)
}
