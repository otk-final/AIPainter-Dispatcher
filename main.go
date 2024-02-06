package main

import (
	"AIPainter-Dispatcher/internal"
	"AIPainter-Dispatcher/internal/ws"
	"context"
	"flag"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"
)

var router = mux.NewRouter()

var address = flag.String("address", ":18181", "Address")

var redisConf = flag.String("redis", "redis://:@localhost:6379/8", "Redis")

var services = flag.String("comfyui_service", "http://58.49.141.134:8188", "ComfyUI服务地址")

var inputAssetPath = flag.String("comfyui_input_path", ".", "ComfyUI上传文件存储路径")

var outputAssetPath = flag.String("comfyui_output_path", ".", "ComfyUI下载文件路径")

var bytedance = flag.String("bytedance_service", "https://openspeech.bytedance.com", "字节服务地址")

var bytedanceToken = flag.String("bytedance_token", "9gyYDsIV-NcEcsbsmErHWK39T9Uvb8Bf", "字节服务TOKEN")

func main() {
	flag.Parse()

	// init redis
	redisOps, err := redis.ParseURL(*redisConf)
	if err != nil {
		log.Panicln(err)
	}
	rdb := redis.NewClient(redisOps)
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Panicln(err)
	}

	lb := internal.NewLoadBalancer(strings.Split(*services, " ")...)
	comfyUIProxy := internal.NewComfyUIProxy("/comfyui", lb, rdb, *inputAssetPath, *outputAssetPath)
	middle := internal.NewMiddleware(rdb)

	//ComfyUI 代理
	comfyUiRouter := router.PathPrefix(comfyUIProxy.PathPrefix).Subrouter()
	//init api router
	comfyUiRouter.Path("/prompt").Methods("POST").HandlerFunc(comfyUIProxy.ApiPrompt)
	//上传到本地，目标服务器通过分布式文件系统读取
	comfyUiRouter.Path("/upload/image").Methods("POST").HandlerFunc(comfyUIProxy.ApiUpload)
	//查询本地目录
	comfyUiRouter.Path("/view").Methods("GET").HandlerFunc(comfyUIProxy.ApiDownload)
	//查询本地记录，不存在则调用目标服务器查询
	comfyUiRouter.Path("/history/{prompt_id}").Methods("GET").HandlerFunc(comfyUIProxy.ApiHistory)
	//长链接
	comfyUiRouter.Path("/ws").HandlerFunc(ws.NewUpgrade)

	//火山引擎 代理
	bytedanceProxy := internal.NewBytedanceManager("/bytedance", *bytedance, *bytedanceToken)
	router.PathPrefix(bytedanceProxy.PathPrefix).Handler(bytedanceProxy.Proxy)

	//cors and auth
	router.Use(mux.CORSMethodMiddleware(router), middle.UserAuthenticationMiddleware)
	//start
	_ = http.ListenAndServe(*address, router)
}
