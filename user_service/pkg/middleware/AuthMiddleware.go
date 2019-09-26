package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
	"strings"
	"time"
)

const(
	TokenExpireErr = "Token expired please login again"
	UserIDContextKey= "userID"
	)

func AuthMiddleware(tracer opentracing.Tracer) gin.HandlerFunc{

	return func(c *gin.Context){
		opentracing.SetGlobalTracer(tracer)
		span, newCtx := opentracing.StartSpanFromContext(c, "user_authentication")
		span.SetBaggageItem("requestID", uuid.New().String())
		span.SetBaggageItem("requestUrl",c.Request.URL.String())
		span.SetTag("funcName","AuthMiddleware")
		span.SetTag("serviceName",tracing.ServiceName)
		span.SetTag("startTime",time.Now().String())
		defer span.Finish()

		jwtKey:=[]byte("SecretKey")
		tokenStr:=c.Request.Header.Get("token")
		claims:=&models.Claims{}
		tkn,err:=jwt.ParseWithClaims(tokenStr,claims,func(token *jwt.Token)(interface{},error){
			return jwtKey,nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			if strings.Contains(err.Error(), "expired") {
				log.Print(err)
				c.JSON(http.StatusUnauthorized,gin.H{
					"msg": nil,
					"error": TokenExpireErr,
				})
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			log.Printf("%v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(UserIDContextKey,claims.ID)
		c.Set("context",newCtx)
		c.Next()
	}



}
