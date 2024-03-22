package middleware

import (
	"AIPainter-Dispatcher/conf"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"strings"
	"time"
)

type Limiter struct {
	cache   *expirable.LRU[string, *rate.Limiter]
	router  *mux.Router
	matcher *mux.RouteMatch
	conf    *conf.LimitConf
}

func NewLimiter(conf *conf.LimitConf) *Limiter {
	//路径匹配
	r := mux.NewRouter()
	lo.ForEach(conf.Predicates, func(item string, index int) {
		r.Methods(strings.Split(item, " ")[0]).Path(strings.Split(item, " ")[1])
	})

	//缓存每个用户的访问速率状态
	cache := expirable.NewLRU[string, *rate.Limiter](10000, nil, time.Hour*2)
	return &Limiter{
		router:  r,
		matcher: &mux.RouteMatch{},
		cache:   cache,
		conf:    conf,
	}
}

func (l *Limiter) Handle(next http.Handler) http.Handler {

	//用户级 限流
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if l.router.Match(request, l.matcher) {

			//获取用户级别限流配置
			up := request.Context().Value(UserPrincipalKey).(*UserPrincipal)

			//默认每分钟
			rateTime := lo.Ternary(l.conf.Rate > 0, l.conf.Rate, 60)
			//区分vip 和 guest
			bucketCount := lo.Ternary(strings.EqualFold(up.Type, "vip"), l.conf.VipBucket, l.conf.Bucket)

			//缓存中获取状态
			cacheKey := fmt.Sprintf("%s/%s", up.Type, up.Id)
			limiter, exit := l.cache.Get(cacheKey)
			if !exit {
				limiter = rate.NewLimiter(rate.Every(time.Second*time.Duration(rateTime)), bucketCount)
				l.cache.Add(cacheKey, limiter)
			}

			if !limiter.Allow() {
				log.Printf("[%s:%s] - [%s]接口请求频繁：%s", up.Type, up.Id, request.Header.Get("x-real-ip"), request.RequestURI)
				http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}
