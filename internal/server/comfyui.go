package server

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/lb"
	"AIPainter-Dispatcher/internal/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ComfyUIProxy struct {
}

func NewComfyUIProxy(conf conf.ComfyUIConf) *httputil.ReverseProxy {

	hash := lb.New(100, nil)
	hash.Add(conf.Address...)

	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {

			//根据用户唯一标识，分配一致性地址
			up := request.Context().Value(middleware.UserPrincipalKey).(*middleware.UserPrincipal)
			targetAddress := hash.Get(up.Id)

			target, err := url.Parse(targetAddress)
			if err != nil {
				return
			}

			request.URL.Scheme = target.Scheme
			request.URL.Host = target.Host
			request.URL.Path = strings.TrimPrefix(request.URL.Path, conf.Location)
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
			request.Host = target.Host
		},
	}
}
