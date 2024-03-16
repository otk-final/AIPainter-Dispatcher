package main

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/middleware"
	"AIPainter-Dispatcher/internal/server"
	"flag"
	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var router = mux.NewRouter()
var address = flag.String("listener", ":18080", "http server listener")
var logPath = flag.String("log", "./access.log", "logger path")
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

func initLogger() {
	r := &lumberjack.Logger{
		Filename:   *logPath,
		MaxSize:    5,
		MaxAge:     7,
		MaxBackups: 10,
		LocalTime:  false,
		Compress:   true,
	}
	multiWriter := io.MultiWriter(r, os.Stdout)

	//结构日志
	logger := slog.New(slog.NewJSONHandler(multiWriter, nil))

	slog.SetDefault(logger)

	//标准库日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(multiWriter)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			_ = r.Rotate()
		}
	}()
}

func main() {
	flag.Parse()

	initLogger()

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
