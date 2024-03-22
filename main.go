package main

import (
	"AIPainter-Dispatcher/conf"
	"AIPainter-Dispatcher/internal/middleware"
	"AIPainter-Dispatcher/internal/server"
	"context"
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

// 直接从请求头中获取身份信息
var identityHandle = func(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		newCtx := context.WithValue(request.Context(), middleware.UserPrincipalKey, &middleware.UserPrincipal{
			Id:   request.Header.Get("x-user-id"),
			Name: request.Header.Get("x-user-name"),
			Type: request.Header.Get("x-user-type"),
		})
		handler.ServeHTTP(writer, request.WithContext(newCtx))
	})
}

func main() {
	flag.Parse()

	initLogger()

	setting := initAppConf(*confPath)

	//代理接口

	//绘图
	if setting.ComfyUI != nil {

		cs := router.PathPrefix(setting.ComfyUI.Location).Subrouter()
		cs.PathPrefix("/").Handler(server.NewComfyUIProxy(setting.ComfyUI))
		if setting.ComfyUI.Limit != nil {
			cs.Use(middleware.NewLimiter(setting.ComfyUI.Limit).Handle)
		}
		log.Printf("开启ComfyUI:%s 限流[%v]", setting.ComfyUI.Address, setting.ComfyUI.Limit != nil)
	}

	//火山引擎
	if setting.Bytedance != nil {
		bs := router.PathPrefix(setting.Bytedance.Location).Subrouter()
		bs.PathPrefix("/").Handler(server.NewBytedanceProxy(setting.Bytedance))
		if setting.Bytedance.Limit != nil {
			bs.Use(middleware.NewLimiter(setting.Bytedance.Limit).Handle)
		}
		log.Printf("开启Bytedance:%s 限流[%v]", setting.Bytedance.Address, setting.Bytedance.Limit != nil)
	}

	//百度
	if setting.Baidu != nil {
		log.Printf("开启Baidu:%s", setting.Baidu.Address)
		bds := router.PathPrefix(setting.Baidu.Location).Subrouter()
		bds.PathPrefix("/").Handler(server.NewBaiduProxy(setting.Baidu))
		if setting.Baidu.Limit != nil {
			bds.Use(middleware.NewLimiter(setting.Baidu.Limit).Handle)
		}
		log.Printf("开启Baidu:%s 限流[%v]", setting.Baidu.Address, setting.Baidu.Limit != nil)
	}

	//OpenAI
	if setting.OpenAI != nil {
		log.Printf("开启OpenAI:%s", setting.OpenAI.Address)
		as := router.PathPrefix(setting.OpenAI.Location).Subrouter()
		as.PathPrefix("/").Handler(server.NewOpenAIProxy(setting.OpenAI))
		if setting.OpenAI.Limit != nil {
			as.Use(middleware.NewLimiter(setting.OpenAI.Limit).Handle)
		}
		log.Printf("开启OpenAI:%s 限流[%v]", setting.OpenAI.Address, setting.OpenAI.Limit != nil)
	}

	//认证 添加用户上下文
	if setting.Jwt != nil {
		log.Println("开启JWT身份认证")
		router.Use(middleware.NewAuth(setting.Jwt).Handle)
	} else {
		router.Use(identityHandle)
	}

	if setting.Redis != nil {
		log.Println("开启Redis请求统计")
		router.Use(middleware.NewStatistics(setting.Redis).Handle)
	}

	//跨域
	handle := cors.AllowAll().Handler(router)
	log.Printf("start http server %s", *address)

	//start
	_ = http.ListenAndServe(*address, handle)
}
