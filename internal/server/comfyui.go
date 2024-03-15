package server

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/lb"
	"AIPainter-Dispatcher/internal/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
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

			//只替换地址和路径,参数保留
			t := target.JoinPath(strings.TrimPrefix(request.URL.Path, conf.Location))
			request.URL.Scheme = t.Scheme
			request.URL.Host = t.Host
			request.URL.Path = path.Join("/", t.Path)
			request.Host = t.Host

			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		},
	}
}
