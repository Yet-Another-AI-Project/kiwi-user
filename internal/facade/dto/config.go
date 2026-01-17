package dto

type WxAppConfigRequest struct {
	Scene string `json:"scene"`
	URL   string `json:"url"`
}

// NOTE: 这里的json格式是为了兼容微信小程序的配置, 不符合go的json规范
type WxAppConfigResponse struct {
	AppId     string   `json:"appId"`
	Timestamp int64    `json:"timestamp"`
	NonceStr  string   `json:"nonceStr"`
	Signature string   `json:"signature"`
	JsAPIList []string `json:"jsApiList"`
	Debug     bool     `json:"debug"`
}
