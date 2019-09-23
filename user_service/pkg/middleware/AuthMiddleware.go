package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"log"
	"net/http"
	"strings"
)

const(
	TokenExpireErr = "Token expired please login again"
	UserIDContextKey= "userID"
	)

func AuthMiddleware(c *gin.Context){
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
	c.Next()
}
