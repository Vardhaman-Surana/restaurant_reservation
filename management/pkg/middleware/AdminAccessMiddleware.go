package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"net/http"
)

func AdminAccessOnly(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=Admin && userAuth.Role!=SuperAdmin{
		c.AbortWithStatus(http.StatusUnauthorized)
		//c.Abort()
		return
	}
	c.Next()
}

func SuperAdminAccessOnly(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=SuperAdmin{
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Next()
}