package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
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

func TokenValidator(db database.Database) gin.HandlerFunc{
	return func(c *gin.Context){
		tokenStr:=c.Request.Header.Get("token")
		isValid:=db.VerifyToken(tokenStr)
		if !isValid{
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Abort()
		}
		c.Next()
	}
}

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
				"msg": TokenExpireMessage,
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
	isValid:=IsValidUserType(claims.Role)
	if !isValid{
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userAuth:=&models.UserAuth{
		ID:   claims.ID,
		Role: claims.Role,
	}
	c.Set("userAuth",userAuth)
	c.Next()
}

func IsValidUserType(userType string)bool{
	if userType!=Admin && userType!=SuperAdmin && userType!=Owner{
		return false
	}
	return true
}