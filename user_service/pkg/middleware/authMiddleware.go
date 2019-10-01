package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
	"strings"
)

const(
	TokenExpireErr = "Token expired please login again"
	UserIDContextKey= "userID"
	)

func AuthMiddleware() gin.HandlerFunc{

	return func(c *gin.Context){
		prevContext,_:=c.Get("context")
		prevCtx:=prevContext.(context.Context)
		span,newCtx:=tracing.GetSpanFromContext(prevCtx,"authentication")
		defer span.Finish()

		tags:=tracing.TraceTags{FuncName:"AuthMiddleware",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
		tracing.SetTags(span,tags)
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


