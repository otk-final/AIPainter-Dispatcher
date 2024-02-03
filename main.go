package main

import (
	"AIPainter-Dispatcher/internal"
	"AIPainter-Dispatcher/internal/ws"
	"flag"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
)

var router = mux.NewRouter()

var address = flag.String("address", ":18080", "Address")

var redisConf = flag.String("redisConf", "redis://root:@localhost:6789/8", "Redis")

var services = flag.String("targets", "http://58.49.141.134:8188", "ComfyUI")

func main() {
	flag.Parse()

	// init redis
	redisOps, err := redis.ParseURL(*redisConf)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(redisOps)

	lb := internal.NewLoadBalancer(strings.Split(*services, " ")...)
	cm := internal.NewComfyUIManager(lb, rdb)
	middle := internal.NewMiddleware(rdb)

	//init api router
	router.Path("/prompt").Methods("POST").HandlerFunc(cm.PromptProxy)

	//上传到本地，目标服务器通过分布式文件系统读取
	router.Path("/upload/image").Methods("POST").HandlerFunc(cm.UploadProxy)

	//查询本地目录
	router.Path("/view").Methods("GET").HandlerFunc(cm.DownloadProxy)

	//查询本地记录，不存在则调用目标服务器查询
	router.Path("/history").Methods("POST").HandlerFunc(cm.HistoryProxy)

	router.Path("/ws").HandlerFunc(ws.NewUpgrade)

	//cors and auth
	router.Use(mux.CORSMethodMiddleware(router), middle.UserAuthenticationMiddleware)

	//start
	_ = http.ListenAndServe(*address, router)
}
