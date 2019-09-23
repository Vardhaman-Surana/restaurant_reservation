package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"net/http"
)

type AdminController struct{
	database.Database
}

func NewAdminController(db database.Database)*AdminController{
	ac:=new(AdminController)
	ac.Database=db
	return ac
}

func(a *AdminController)GetAdmins(c *gin.Context){
	jsonData:=&[]models.UserOutput{}
	var stringData string
	var err error
	stringData,err=a.ShowAdmins()
	if err!=nil{
		c.Status(http.StatusInternalServerError)
	}
	if stringData!=""{
		_ = json.Unmarshal([]byte(stringData), jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
}

func(a *AdminController)EditAdmin(c *gin.Context){
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
}

func(a *AdminController)DeleteAdmins(c *gin.Context){
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
}