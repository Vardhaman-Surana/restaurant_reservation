package controller

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"log"
	"net/http"
	"time"
)

type HelloWorldController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewHelloWorldController(db database.Database,logger *fluent.Fluent)*HelloWorldController{
	hc:=new(HelloWorldController)
	hc.Database=db
	hc.Logger=logger
	return hc
}

func(h *HelloWorldController)SayHello(c *gin.Context){
	er:=h.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Hello World",
		"time": time.Now().String(),
		"created_by" : "vardhaman",
		"completed":"23-08-2019",
	})
	er=h.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}