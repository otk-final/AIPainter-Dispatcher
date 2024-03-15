package server

import (
	"AIPainter-Dispatcher/conf"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

func NewOpenAIProxy(conf conf.OpenAIConf) *httputil.ReverseProxy {
	target, err := url.Parse(conf.Address)
	if err != nil {
		log.Panicln(err)
	}

	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {

			//只替换地址和路径,参数保留
			t := target.JoinPath(strings.TrimPrefix(request.URL.Path, conf.Location))
			request.URL.Scheme = t.Scheme
			request.URL.Host = t.Host
			request.URL.Path = path.Join("/", t.Path)
			request.Host = t.Host

			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", conf.Authorization))
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		},
	}
}