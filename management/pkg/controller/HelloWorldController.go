package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"net/http"
	"time"
)

type HelloWorldController struct{
	database.Database
}

func NewHelloWorldController(db database.Database)*HelloWorldController{
	hc:=new(HelloWorldController)
	hc.Database=db
	return hc
}

func(h *HelloWorldController)SayHello(c *gin.Context){
	c.JSON(http.StatusOK,gin.H{
		"msg":"Hello World",
		"time": time.Now().String(),
		"created_by" : "vardhaman",
		"completed":"23-08-2019",
	})
}