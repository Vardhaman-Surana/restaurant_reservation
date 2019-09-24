package controller

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"time"
)

const(
	Success="Success"
	Fail="Fail"
)

type LogInController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewLogInController(db database.Database,logger *fluent.Fluent)*LogInController{
	lc:=new(LogInController)
	lc.Database=db
	lc.Logger=logger
	return lc
}
func(l *LogInController)LogIn(c *gin.Context){
	er:=l.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
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
	er=l.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func(l *LogInController)LogOut(c *gin.Context){
	er:=l.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
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
	er=l.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}


