package controller

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"time"
)

const ErrJsonInput="Invalid Json Input"

type RegisterController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewRegisterController(db database.Database,logger *fluent.Fluent) *RegisterController{
	regController:=new(RegisterController)
	regController.Database=db
	regController.Logger=logger
	return regController
}

func(r *RegisterController)Register(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	var user models.UserReg
	err:=c.ShouldBindJSON(&user)
	if err!=nil {
		fmt.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
			"status":Fail,
		})
		return
	}
	if user.Role!=middleware.Admin && user.Role!=middleware.SuperAdmin {
		c.Status(http.StatusNotFound)
		return
	}
	err=r.CreateUser(&user)
	if err!=nil{
		if err==database.ErrDupEmail{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": database.ErrDupEmail.Error(),
				"status":Fail,
			})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"role":user.Role,
		"msg":"Registration Successful",
		"status":Success,
	})
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}