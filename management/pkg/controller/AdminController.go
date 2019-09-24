package controller

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"runtime"
	"time"
)
const Tag = "restaurant.management"


type AdminController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewAdminController(db database.Database,logger *fluent.Fluent)*AdminController{
	ac:=new(AdminController)
	ac.Database=db
	ac.Logger=logger
	return ac
}

func(a *AdminController)GetAdmins(c *gin.Context){
	er:=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	jsonData:=&[]models.UserOutput{}
	var stringData string
	var err error
	stringData,err=a.ShowAdmins()
	if err!=nil{
		c.Status(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_ = json.Unmarshal([]byte(stringData), jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
	er=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func(a *AdminController)EditAdmin(c *gin.Context){
	er:=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	adminID := c.Param("adminID")
	var admin models.UserOutput
	admin.ID=adminID
	err:=c.ShouldBindJSON(&admin)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err = a.CheckAdmin(admin.ID)
	if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":"Admin does not exist",
			})
			return
	}
	err=a.UpdateAdmin(&admin)
	if err!=nil{
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"admin updated successfully",
	})
	er=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func(a *AdminController)DeleteAdmins(c *gin.Context){
	er:=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	var adminID struct {
		IDArr []string	`json:"idArr" binding:"required"`
	}
	err:=c.ShouldBindJSON(&adminID)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=a.RemoveAdmins(adminID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Admins deleted successfully",
	})
	er=a.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func GetfuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}