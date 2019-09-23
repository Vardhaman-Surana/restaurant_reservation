package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"net/http"
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
