package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
	"time"
)

type RestaurantController struct{
	db database.Database
	tracer opentracing.Tracer
}

func NewRestaurantController(db database.Database,tracer opentracing.Tracer)*RestaurantController{
	resc:=new(RestaurantController)
	resc.db=db
	resc.tracer=tracer
	return resc
}

func(resc *RestaurantController)GetRestaurants(c *gin.Context){
	prevContext,_:=c.Get("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx := opentracing.StartSpanFromContext(prevCtx,"user_get_restaurants")
	span.SetTag("serviceName",tracing.ServiceName)
	span.SetTag("startTime",time.Now().String())
	defer span.Finish()

	restaurants,err:=resc.db.GetRestaurants(newCtx)
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