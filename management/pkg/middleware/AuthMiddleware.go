package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"log"
	"net/http"
	"strings"
)

const(
	Admin="admin"
 	SuperAdmin="superAdmin"
 	Owner="owner"
	TokenExpireMessage="Token expired please login again"
)

func TokenValidator(db database.Database) mux.MiddlewareFunc{
	return func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request){
		prevContext:=rq.Context().Value("context")
		prevCtx:=prevContext.(context.Context)
		span,newCtx:=tracing.GetSpanFromContext(prevCtx,"token_validation")
		defer span.Finish()
		tags:=tracing.TraceTags{FuncName:"TokenValidator",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
		tracing.SetTags(span,tags)

		tokenStr:=rq.Header.Get("token")
		vfTknCtx,isValid:=db.VerifyToken(newCtx,tokenStr)
		if !isValid{
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reqCtx:=context.WithValue(rq.Context(),"context",vfTknCtx)
		next.ServeHTTP(w,rq.WithContext(reqCtx))
	})}
}

func AuthMiddleware(next http.Handler) http.Handler{
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
			log.Print(err)
			w.WriteHeader(http.StatusUnauthorized)
			bodyMap:=models.DefaultMap{
				"msg": TokenExpireMessage,
			}
			w.Write(bodyMap.ConvertToByteArray())
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
	isValid:=IsValidUserType(claims.Role)
	if !isValid{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userAuth:=&models.UserAuth{
		ID:   claims.ID,
		Role: claims.Role,
	}
	reqCtx:=context.WithValue(rq.Context(),"context",newCtx)
	reqCtx=context.WithValue(reqCtx,"userAuth",userAuth)

	next.ServeHTTP(w,rq.WithContext(reqCtx))
	})
}

func IsValidUserType(userType string)bool{
	if userType!=Admin && userType!=SuperAdmin && userType!=Owner{
		return false
	}
	return true
}