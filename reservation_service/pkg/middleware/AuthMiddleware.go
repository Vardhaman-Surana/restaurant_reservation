package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/models"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/tracing"
	"log"
	"net/http"
	"strings"
)

const(
	TokenExpireErr = "Token expired please login again"
	UserIDContextKey= "userID"
)

func AuthMiddleware(next http.Handler)http.Handler{

	return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request){

	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"authentication")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AuthMiddleware",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	jwtKey:=[]byte("SecretKey")
	tokenStr:=rq.Header.Get("token")
	claims:=&models.Claims{}
	tkn,err:=jwt.ParseWithClaims(tokenStr,claims,func(token *jwt.Token)(interface{},error){
		return jwtKey,nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "expired") {
			log.Println(err)
			models.WriteToResponse(w,http.StatusUnauthorized,&models.DefaultMap{
				"msg": nil,
				"error": TokenExpireErr,
			})
			return
		}
		log.Printf("%v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	reqContext:=context.WithValue(rq.Context(),"context",newCtx)
	reqContext=context.WithValue(reqContext,UserIDContextKey,claims.ID)
	next.ServeHTTP(w,rq.WithContext(reqContext))
})
}
