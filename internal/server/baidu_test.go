package server

import (
	"AIPainter-Dispatcher/conf"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestTokenLoad(t *testing.T) {
	tokenFile := "../../conf/baidu_auth.json"
	auth2, err := tokenLoad(tokenFile)
	if err != nil {
		t.Error(err)
	}
	auth2.RefreshTime = time.Now()
	auth2.ExpiredTime = time.Now().Add(time.Second * time.Duration(auth2.ExpiresIn-3600))

	//存储本地文件
	saveBytes, _ := json.Marshal(auth2)
	_ = os.WriteFile(tokenFile, saveBytes, os.ModePerm)

	//t.Log(a)
}

func TestTokenNexDuration(t *testing.T) {
	tokenFile := "../../conf/baidu_auth.json"
	auth2, err := tokenLoad(tokenFile)
	if err != nil {
		t.Error(err)
	}
	t.Log(auth2.NexDuration())
}

func TestTokenRefreshJob(t *testing.T) {
	target, _ := url.Parse("https://aip.baidubce.com")
	tokenRefreshJob(target, conf.BaiduConf{
		Location:     "",
		Address:      "https://aip.baidubce.com",
		ClientId:     "j5YmC3RxnEay92YGDzKfepP8",
		ClientSecret: "IlRssYk8nswWfRE3Yy81GqPsZKmLkdAv",
	})
}
