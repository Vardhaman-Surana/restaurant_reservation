package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
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
	prevContext,_:=c.Get("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx :=tracing.GetSpanFromContext(prevCtx,"get_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	_,restaurants,err:=resc.db.SelectRestaurants(newCtx)
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