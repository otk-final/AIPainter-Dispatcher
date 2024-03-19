package server

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/lb"
	"AIPainter-Dispatcher/internal/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

//type ComfyUIProxy struct {
//	address []string
//	lock    sync.Mutex
//}
//
//func (c *ComfyUIProxy) promptQueueCheck(targetAddress string) (int, error) {
//	target, _ := url.Parse(targetAddress)
//	resp, err := http.Get(target.JoinPath("/prompt").String())
//	if err != nil {
//		return -1, nil
//	}
//
//	defer resp.Body.Close()
//	body, _ := io.ReadAll(resp.Body)
//	var inf map[string]any
//	_ = json.Unmarshal(body, &inf)
//
//	//获取当前队列剩余任务数
//	return inf["exec_info"].(map[string]any)["queue_remaining"].(int), nil
//}

func NewComfyUIProxy(conf *conf.ComfyUIConf) *httputil.ReverseProxy {

	hash := lb.New(100, nil)
	hash.Add(conf.Address...)

	//开启异步监听各个实力队列情况
	//cache := expirable.NewLRU[string, *UserLastedJob](1000, nil, time.Minute*3)

	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {

			//根据用户唯一标识，分配一致性地址
			up := request.Context().Value(middleware.UserPrincipalKey).(*middleware.UserPrincipal)

			//接口一致性要求高 文件上传，提交任务，查询结果，需落地到同一服务器
			//根据请求头 x-user-id 和 x-trace-id 计算一致性
			traceId := request.Header.Get("x-trace-id")

			//获取缓存中的历史记录
			key := strings.Join([]string{up.Id, traceId}, "#")
			targetAddress := hash.Get(key)
			target, _ := url.Parse(targetAddress)

			//只替换地址和路径,参数保留
			t := target.JoinPath(strings.TrimPrefix(request.URL.Path, conf.Location))
			request.URL.Scheme = t.Scheme
			request.URL.Host = t.Host
			request.URL.Path = path.Join("/", t.Path)
			request.Host = t.Host
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		},
		ModifyResponse: func(response *http.Response) error {
			response.Header.Del("Access-Control-Allow-Origin")
			return nil
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Printf("ComfyUIProxyError: %s => %s %s", request.RequestURI, request.Host, err.Error())
			writer.WriteHeader(http.StatusBadGateway)
		},
	}
}
