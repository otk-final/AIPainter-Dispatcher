package middleware

import (
	"AIPainter-Dispatcher/conf"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type UserPrincipal struct {
	Id     string
	Name   string
	Type   string
	Claims *UserClaims
}

type UserClaims struct {
	jwt.Claims
	Vip           string
	VipRegTime    string
	VipRenewTime  string
	VipExpireTime string
}

const UserPrincipalKey = "UserPrincipal"

type Auth struct {
	jwtParser jwt.Parser
	jwtConf   conf.JwtConf
}

func NewAuth(jwtConf conf.JwtConf) *Auth {
	return &Auth{jwtConf: jwtConf}
}

func (a *Auth) HandleMock(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		newCtx := context.WithValue(request.Context(), UserPrincipalKey, &UserPrincipal{
			Id:   "abc",
			Name: "测试",
			Type: "vip",
			//Claims: token.Claims.(*UserClaims),
		})
		handler.ServeHTTP(writer, request.WithContext(newCtx))
	})
}

func (a *Auth) Handle(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if authorization == "" {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}

		//jwt 解码
		token, err := a.jwtParser.ParseWithClaims(authorization, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtConf.Key), nil
		})

		if err != nil {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}

		newCtx := context.WithValue(request.Context(), UserPrincipalKey, &UserPrincipal{
			Id:     token.Header["x-user-id"].(string),
			Name:   token.Header["x-user-name"].(string),
			Type:   token.Header["x-user-type"].(string),
			Claims: token.Claims.(*UserClaims),
		})
		handler.ServeHTTP(writer, request.WithContext(newCtx))
	})
}
