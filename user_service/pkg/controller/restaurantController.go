package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"log"
	"net/http"
)

type RestaurantController struct{
	db database.Database
}

func NewRestaurantController(db database.Database)*RestaurantController{
	resc:=new(RestaurantController)
	resc.db=db
	return resc
}

func(resc *RestaurantController)GetRestaurants(c *gin.Context){
	restaurants,err:=resc.db.GetRestaurants()
	if err!=nil{
		log.Printf("error retrieving restaurants:%v",err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	c.JSON(http.StatusOK,restaurants)
}