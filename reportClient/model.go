package reportClient

type HealthReportForm struct {
	Token                          string `json:"_token"`
	JibenxinxiShifoubenrenshangbao string `json:"jibenxinxi_shifoubenrenshangbao"`
	Profile                        struct {
		Xuegonghao  string `json:"xuegonghao"`
		Xingming    string `json:"xingming"`
		Suoshubanji string `json:"suoshubanji"`
	} `json:"profile"`
	JiankangxinxiMuqianshentizhuangkuang    string        `json:"jiankangxinxi_muqianshentizhuangkuang"`
	XingchengxinxiWeizhishifouyoubianhua    string        `json:"xingchengxinxi_weizhishifouyoubianhua"`
	CrossCity                               string        `json:"cross_city"`
	QitashixiangQitaxuyaoshuomingdeshixiang string        `json:"qitashixiang_qitaxuyaoshuomingdeshixiang"`
	Credits                                 string        `json:"credits"`
	BmapPosition                            string        `json:"bmap_position"`
	BmapPositionLatitude                    string        `json:"bmap_position_latitude"`
	BmapPositionLongitude                   string        `json:"bmap_position_longitude"`
	BmapPositionAddress                     string        `json:"bmap_position_address"`
	BmapPositionStatus                      string        `json:"bmap_position_status"`
	ProvinceCode                            string        `json:"ProvinceCode"`
	CityCode                                string        `json:"CityCode"`
	Travels                                 []interface{} `json:"travels"`
}

type StudentInfo struct {
	Data struct {
		Xuegonghao             string      `json:"xuegonghao"`
		Xingming               string      `json:"xingming"`
		Suoshudanwei           string      `json:"suoshudanwei"`
		Xingbie                string      `json:"xingbie"`
		Lianxidianhua          string      `json:"lianxidianhua"`
		Zhengjianhaoma         string      `json:"zhengjianhaoma"`
		Chushengriqi           string      `json:"chushengriqi"`
		Zhengjianleixing       string      `json:"zhengjianleixing"`
		Shenfenleixing         string      `json:"shenfenleixing"`
		Suoshubanji            string      `json:"suoshubanji"`
		Zhiwu                  interface{} `json:"zhiwu"`
		Jinjilianxirenxingming string      `json:"jinjilianxirenxingming"`
		Jinjilianxirendianhua  string      `json:"jinjilianxirendianhua"`
		Credits                int         `json:"credits"`
		Wuhan                  struct {
			Data struct {
				Id int `json:"id"`
			} `json:"data"`
		} `json:"wuhan"`
		Notes struct {
			Data []struct {
				CreatedOn                  string      `json:"created_on"`
				AddressArea                interface{} `json:"address_area"`
				XingchengxinxiGuojia       string      `json:"xingchengxinxi_guojia"`
				XingchengxinxiShengfen     string      `json:"xingchengxinxi_shengfen"`
				XingchengxinxiChengshi     string      `json:"xingchengxinxi_chengshi"`
				XingchengxinxiQuxian       interface{} `json:"xingchengxinxi_quxian"`
				XingchengxinxiXiangxidizhi interface{} `json:"xingchengxinxi_xiangxidizhi"`
				Credits                    int         `json:"credits"`
			} `json:"data"`
		} `json:"notes"`
	} `json:"data"`
}
