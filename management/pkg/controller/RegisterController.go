package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"net/http"
)

const ErrJsonInput="Invalid Json Input"

type RegisterController struct{
	database.Database
}

func NewRegisterController(db database.Database) *RegisterController{
	regController:=new(RegisterController)
	regController.Database=db
	return regController
}

func(r *RegisterController)Register(c *gin.Context){
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
}