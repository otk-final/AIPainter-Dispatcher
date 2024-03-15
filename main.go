package main

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/middleware"
	"AIPainter-Dispatcher/internal/server"
	"flag"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

var router = mux.NewRouter()
var address = flag.String("listener", ":18080", "http server listener")
var confPath = flag.String("conf", "conf/conf.yaml", "config path")

func initAppConf(confPath string) *conf.AppConfig {

	//文件
	viper.SetConfigFile(confPath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panicln(err)
	}
	var config conf.AppConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Panicln(err)
	}
	return &config
}

func main() {

	flag.Parse()
	setting := initAppConf(*confPath)

	//代理接口

	//绘图
	cs := router.PathPrefix(setting.ComfyUI.Location).Subrouter()
	cs.PathPrefix("/").Handler(server.NewComfyUIProxy(setting.ComfyUI))
	cs.Use(middleware.NewLimiter(setting.ComfyUI.Limit).Handle)

	//火山引擎
	bs := router.PathPrefix(setting.Bytedance.Location).Subrouter()
	bs.PathPrefix("/").Handler(server.NewBytedanceProxy(setting.Bytedance))
	bs.Use(middleware.NewLimiter(setting.Bytedance.Limit).Handle)

	//百度
	bds := router.PathPrefix(setting.Baidu.Location).Subrouter()
	bds.PathPrefix("/").Handler(server.NewBaiduProxy(setting.Baidu))
	bds.Use(middleware.NewLimiter(setting.Baidu.Limit).Handle)

	//OpenAI
	as := router.PathPrefix(setting.OpenAI.Location).Subrouter()
	as.PathPrefix("/").Handler(server.NewOpenAIProxy(setting.OpenAI))
	as.Use(middleware.NewLimiter(setting.OpenAI.Limit).Handle)

	//认证 + 统计
	router.Use(middleware.NewAuth(setting.Jwt).Handle, middleware.NewStatistics(setting.Redis).Handle)

	//跨域
	handle := cors.AllowAll().Handler(router)
	log.Printf("start http server %s", *address)

	//start
	_ = http.ListenAndServe(*address, handle)
}
