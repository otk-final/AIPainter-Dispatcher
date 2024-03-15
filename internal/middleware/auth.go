package middleware

import (
	"AIPainter-Dispatcher/conf"
	"context"
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
)

type UserPrincipal struct {
	Id   string
	Name string
	Type string
}

const UserPrincipalKey = "UserPrincipal"

type Auth struct {
	rasKey  *rsa.PublicKey
	jwtConf conf.JwtConf
}

func NewAuth(jwtConf conf.JwtConf) *Auth {

	//读取公钥
	publicKeyData, err := os.ReadFile(jwtConf.PublicKey)
	pk, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		log.Panicln(err)
	}

	return &Auth{jwtConf: jwtConf, rasKey: pk}
}

func (a *Auth) Handle(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if authorization == "" {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}

		//jwt 解码
		token, err := jwt.ParseWithClaims(strings.Split(authorization, " ")[1], &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return a.rasKey, nil
		})
		if err != nil {
			http.Error(writer, "Client Unauthorized", http.StatusUnauthorized)
			return
		}
		//mapClaims := token.Claims.(*jwt.MapClaims)
		newCtx := context.WithValue(request.Context(), UserPrincipalKey, &UserPrincipal{
			Id:   token.Header["x-user-id"].(string),
			Name: token.Header["x-user-name"].(string),
			Type: token.Header["x-user-type"].(string),
		})
		handler.ServeHTTP(writer, request.WithContext(newCtx))
	})
}
