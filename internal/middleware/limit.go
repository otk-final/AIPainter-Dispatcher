package middleware

import (
	"AIPainter-Dispatcher/conf"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Limiter struct {
	cache   sync.Map
	router  *mux.Router
	matcher *mux.RouteMatch
	conf    conf.LimitConf
}

func NewLimiter(conf conf.LimitConf) *Limiter {

	//路径匹配
	r := mux.NewRouter()
	lo.ForEach(conf.Predicates, func(item string, index int) {
		r.Methods(strings.Split(item, " ")[0]).Path(strings.Split(item, " ")[1])
	})

	return &Limiter{
		router:  r,
		matcher: &mux.RouteMatch{},
		cache:   sync.Map{},
		conf:    conf,
	}
}

func (l *Limiter) Handle(next http.Handler) http.Handler {

	//用户级 限流
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if l.router.Match(request, l.matcher) {
			//获取用户级别限流配置
			up := request.Context().Value(UserPrincipalKey).(*UserPrincipal)

			rateTime := lo.Ternary(strings.EqualFold(up.Type, "vip"), l.conf.VipRate, l.conf.Rate)
			bucketCount := lo.Ternary(strings.EqualFold(up.Type, "vip"), l.conf.VipBucket, l.conf.Bucket)

			//区分vip 和 guest
			data, _ := l.cache.LoadOrStore(fmt.Sprintf("%s/%s", up.Type, up.Id), rate.NewLimiter(rate.Every(time.Second*time.Duration(rateTime)), bucketCount))
			limiter := data.(*rate.Limiter)

			if !limiter.Allow() {
				log.Printf("[%s]接口请求频繁：%s - %f", up.Id, request.RequestURI, limiter.Tokens())
				http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}
