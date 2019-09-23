package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
)

const(
	Success="Success"
	Fail="Fail"
)

type LogInController struct{
	database.Database
}

func NewLogInController(db database.Database)*LogInController{
	lc:=new(LogInController)
	lc.Database=db
	return lc
}
func(l *LogInController)LogIn(c *gin.Context){
	var cred models.Credentials
	err:=c.ShouldBindJSON(&cred)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}

	isValid:=middleware.IsValidUserType(cred.Role)
	if !isValid{
		c.Status(http.StatusNotFound)
		return
	}
	userID,err:=l.LogInUser(&cred)
	if err!=nil{
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": database.ErrInvalidCredentials.Error(),
			"status": Fail,
		})
		return
	}
	claims:=&models.Claims{
		ID:userID,
		Role:cred.Role,
	}
	token,err:=encryption.CreateToken(claims)
	if err!=nil{
		log.Printf("%v",err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.Header().Set("token",token)
	c.JSON(http.StatusOK,gin.H{
		"role":cred.Role,
		"msg":"Login Successful",
		"status":Success,
	})
}

func(l *LogInController)LogOut(c *gin.Context){
	tokenStr:=c.Request.Header.Get("token")
	if tokenStr==""{
		c.Status(http.StatusBadRequest)
		return
	}
	err:=l.StoreToken(tokenStr)
	if err!=nil{
		c.Status(http.StatusInternalServerError)
		return
	}
	go l.DeleteExpiredToken(tokenStr,encryption.ExpireDuration)
	c.JSON(http.StatusOK,gin.H{
		"msg":"Logged Out Successfully",
		"status":Success,
	})
}


