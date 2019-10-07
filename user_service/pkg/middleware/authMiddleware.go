package middleware

import (
	"context"
	"fmt"
	"github.com/vds/restaurant_reservation/user_service/pkg/fireBaseAuth"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
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
		//jwtKey:=[]byte("SecretKey")
		tokenStr:=rq.Header.Get("Authorization")
		/*claims:=&models.Claims{}
		tkn,err:=jwt.ParseWithClaims(tokenStr,claims,func(token *jwt.Token)(interface{},error){
			return jwtKey,nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if strings.Contains(err.Error(), "expired") {
				log.Print(err)
				models.WriteToResponse(w,http.StatusUnauthorized,&models.DefaultMap{
					"msg": nil,
					"error": TokenExpireErr,
				})
			}
			log.Printf("%v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}*/
		fmt.Printf("token is %v", tokenStr)
		idtkn,err:=fireBaseAuth.SignInWithCustomToken(tokenStr)

		if err!=nil{
			log.Printf("error generating id token:%v",err)
			return
		}


		userID,err:=fireBaseAuth.VerifyToken(newCtx,idtkn)
		if err!=nil{
			log.Print(err)
			models.WriteToResponse(w,http.StatusUnauthorized,&models.DefaultMap{
				"msg": nil,
				"error": TokenExpireErr,
			})
			return
		}


		reqContext:=context.WithValue(rq.Context(),"context",newCtx)
		reqContext=context.WithValue(reqContext,UserIDContextKey,userID)
		next.ServeHTTP(w,rq.WithContext(reqContext))
	})
}


