package internal

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func withBytedanceProxy(pathPrefix string, target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = target.Scheme
			request.URL.Host = target.Host
			request.URL.Path = strings.TrimPrefix(request.URL.Path, pathPrefix)
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
			request.Host = target.Host
		},
		Transport:     &http.Transport{},
		FlushInterval: 0,
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			slog.Warn("Proxy err")
		},
	}
}

type BytedanceManager struct {
	PathPrefix string
	Target     string
	Proxy      *httputil.ReverseProxy
}

func NewBytedanceManager(pathPrefix, host, token string) *BytedanceManager {

	target, err := url.Parse(host)
	if err != nil {
		log.Panicln(err)
	}

	return &BytedanceManager{
		PathPrefix: pathPrefix,
		Proxy: &httputil.ReverseProxy{
			Director: func(request *http.Request) {
				request.URL.Scheme = target.Scheme
				request.URL.Host = target.Host
				request.URL.Path = strings.TrimPrefix(request.URL.Path, pathPrefix)
				request.Header.Set("Authorization", fmt.Sprintf("Bearer;%s", token))
				request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
				request.Host = target.Host
			},
			Transport:     &http.Transport{},
			FlushInterval: 0,
			ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
				slog.Warn("bytedance Proxy err", err)
			},
		},
	}
}
