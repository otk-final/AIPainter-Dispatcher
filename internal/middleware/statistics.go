package middleware

import (
	"AIPainter-Dispatcher/conf"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
)

type Statistics struct {
	conf conf.RedisConf
	rdb  *redis.Client
}

func NewStatistics(conf conf.RedisConf) *Statistics {

	//init redis
	redisOps, err := redis.ParseURL(conf.Address)
	if err != nil {
		log.Panicln(err)
	}
	rdb := redis.NewClient(redisOps)
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Panicln(err)
	}

	return &Statistics{
		rdb:  rdb,
		conf: conf,
	}
}

func (s *Statistics) Handle(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		rdbCtx := context.Background()

		up := request.Context().Value(UserPrincipalKey).(*UserPrincipal)

		//ip访问
		realIp := request.Header.Get("X-Real-IP")
		if realIp != "" {
			s.rdb.PFAdd(rdbCtx, "ip", realIp)
		}

		//用户
		s.rdb.PFAdd(rdbCtx, "user", up.Id)
		s.rdb.PFAdd(rdbCtx, fmt.Sprintf("uri:%s", request.RequestURI), up.Id)

		next.ServeHTTP(writer, request)
	})
}
