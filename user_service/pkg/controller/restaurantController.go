package controller

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"log"
	"net/http"
	"runtime"
	"time"
)
const Tag = "restaurant.user"


type RestaurantController struct{
	db database.Database
	Logger *fluent.Fluent
}

func NewRestaurantController(db database.Database,logger *fluent.Fluent)*RestaurantController{
	resc:=new(RestaurantController)
	resc.db=db
	resc.Logger=logger
	return resc
}

func(resc *RestaurantController)GetRestaurants(c *gin.Context){
	er:=resc.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	restaurants,err:=resc.db.GetRestaurants()
	if err!=nil{
		log.Printf("error retrieving restaurants:%v",err)
		er:=resc.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":"","info":fmt.Sprintf("err is %v",err)})
		if er!=nil{
			log.Printf("error in posting log:%v",er)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	c.JSON(http.StatusOK,restaurants)
	er=resc.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func GetfuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}