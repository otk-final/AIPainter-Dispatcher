package server

import (
	"AIPainter-Dispatcher/conf"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type auth2Token struct {
	AccessToken string `json:"access_token"`

	//剩余秒
	ExpiresIn     int64  `json:"expires_in"`
	RefreshToken  string `json:"refresh_token"`
	Scope         string `json:"scope"`
	SessionKey    string `json:"session_key"`
	SessionSecret string `json:"session_secret"`

	//自定义
	RefreshTime time.Time `json:"refresh_time"`
	ExpiredTime time.Time `json:"expired_time"`
}

func (t *auth2Token) IsExpired() bool {
	return t.ExpiredTime.Before(time.Now())
}

func (t *auth2Token) NexDuration() time.Duration {
	return t.ExpiredTime.Sub(time.Now())
}

var token *auth2Token

// 刷新 access_token
func tokenRefreshJob(target *url.URL, conf conf.BaiduConf) {

	//tokenFile := "../../conf/baidu_auth.json"

	//读取本地文件
	existToken, err := tokenLoad(conf.TokenFile)
	var nextDuration time.Duration

	if os.IsNotExist(err) || existToken.IsExpired() {
		//不存在 或者已经过期 3秒后查询
		nextDuration = time.Second * 3
	} else {
		token = existToken
		nextDuration = existToken.NexDuration()
	}

	//开启定时
	timer := time.NewTimer(nextDuration)
	defer timer.Stop()

	for {
		<-timer.C

		//查询
		at, err := tokenHandle(target, conf, conf.TokenFile)
		if err != nil {
			log.Println("token handle err:{}", err.Error())
			continue
		}

		token = at

		//重置下次触发时间
		timer.Reset(at.NexDuration())
	}
}

// 读取本地token
func tokenLoad(file string) (*auth2Token, error) {
	//读取本地文件
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fb, _ := io.ReadAll(f)
	var auth2 auth2Token
	err = json.Unmarshal(fb, &auth2)
	if err != nil {
		return nil, err
	}
	return &auth2, nil
}

// 获取 access_token
func tokenHandle(target *url.URL, conf conf.BaiduConf, file string) (*auth2Token, error) {

	raws := url.Values{}
	raws.Set("client_id", conf.ClientId)
	raws.Set("client_secret", conf.ClientSecret)
	raws.Set("grant_type", "client_credentials")

	tokenURL := &url.URL{
		Scheme: target.Scheme,
		Host:   target.Host,
		Path:   "/oauth/2.0/token",
	}

	//调用
	resp, err := http.PostForm(tokenURL.String(), raws)
	if err != nil {
		log.Println("baidu token refresh failed {}", err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	//解码
	var auth2 auth2Token
	err = json.Unmarshal(body, &auth2)
	if err != nil {
		return nil, err
	}

	//记录刷新时间和过期时间(提前一个小时）
	auth2.RefreshTime = time.Now()
	auth2.ExpiredTime = time.Now().Add(time.Second * time.Duration(auth2.ExpiresIn-3600))

	//存储本地文件
	saveBytes, _ := json.Marshal(auth2)
	_ = os.WriteFile(file, saveBytes, os.ModePerm)

	return &auth2, nil
}

func NewBaiduProxy(conf conf.BaiduConf) *httputil.ReverseProxy {
	target, err := url.Parse(conf.Address)
	if err != nil {
		log.Panicln(err)
	}

	//开启独立线程监听
	go tokenRefreshJob(target, conf)

	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			//携带token
			rawValues := url.Values{}
			rawValues.Set("access_token", token.AccessToken)

			//只替换地址和路径,参数保留
			t := target.JoinPath(strings.TrimPrefix(request.URL.Path, conf.Location))
			request.URL.Scheme = t.Scheme
			request.URL.Host = t.Host
			request.URL.Path = path.Join("/", t.Path)
			request.URL.RawQuery = rawValues.Encode()
			request.Host = t.Host

			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		},
		ModifyResponse: func(response *http.Response) error {
			response.Header.Del("Access-Control-Allow-Origin")
			return nil
		},
	}
}
