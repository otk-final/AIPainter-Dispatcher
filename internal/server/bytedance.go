package server

import (
	"AIPainter-Dispatcher/conf"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewBytedanceProxy(conf conf.BytedanceConf) *httputil.ReverseProxy {
	target, err := url.Parse(conf.Address)
	if err != nil {
		log.Panicln(err)
	}
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = target.Scheme
			request.URL.Host = target.Host
			request.URL.Path = strings.TrimPrefix(request.URL.Path, conf.Location)
			request.Header.Set("Authorization", fmt.Sprintf("Bearer;%s", conf.Authorization))
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
			request.Host = target.Host
		},
	}
}
