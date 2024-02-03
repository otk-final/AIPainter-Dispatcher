package internal

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type Middleware struct {
	rdb *redis.Client
	jp  *jwt.Parser
}

func NewMiddleware(rdb *redis.Client) *Middleware {
	return &Middleware{rdb: rdb, jp: jwt.NewParser()}
}

func (m *Middleware) UserAuthenticationMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		authorization := request.Header.Get("Authorization")
		if authorization == "" {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}

		//jwt 解码
		token, err := m.jp.Parse(authorization, nil)
		if err != nil {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}

		request.WithContext(context.WithValue(request.Context(), UserPrincipalKey, &UserPrincipal{
			UserId:   token.Header["userId"].(string),
			UserType: token.Header["userType"].(string),
		}))
		handler.ServeHTTP(writer, request)
	})
}
