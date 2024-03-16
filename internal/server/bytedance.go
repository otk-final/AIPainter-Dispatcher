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

func NewBytedanceProxy(conf conf.BytedanceConf) *httputil.ReverseProxy {
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

			//添加认证信息
			request.Header.Set("Authorization", fmt.Sprintf("Bearer;%s", conf.Authorization))
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		},
		ModifyResponse: func(response *http.Response) error {
			response.Header.Del("Access-Control-Allow-Origin")
			return nil
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Printf("BytedanceProxyError: %s => %s %s", request.RequestURI, request.Host, err.Error())
			writer.WriteHeader(http.StatusBadGateway)
		},
	}
}
