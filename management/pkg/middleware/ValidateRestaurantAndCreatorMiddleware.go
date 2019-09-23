package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"net/http"
	"strconv"
)

func ValidateRestaurantAndCreator(db database.Database) gin.HandlerFunc{
	return func(c *gin.Context) {
		value, _ := c.Get("userAuth")
		userAuth := value.(*models.UserAuth)
		res := c.Param("resID")
		resID, _ := strconv.Atoi(res)
		if userAuth.Role == Admin {
			err := db.CheckRestaurantCreator(userAuth.ID, resID)
			if err != nil {
				if err != database.ErrInternal {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": err.Error(),
					})
					c.Abort()
					return
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}else if userAuth.Role == Owner {
			err := db.CheckRestaurantOwner(userAuth.ID, resID)
			if err != nil {
				if err != database.ErrInternal {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": err.Error(),
					})
					c.Abort()
					return
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
		c.Set("restaurantID",resID)
		c.Next()
	}
}
