package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"net/http"
)

type OwnerController struct{
	database.Database
}

func NewOwnerController(db database.Database)*OwnerController{
	ownerController:=new(OwnerController)
	ownerController.Database=db
	return ownerController
}
func(o *OwnerController)GetOwners(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.UserOutput{}
	var stringData string
	var err error
	stringData,err=o.ShowOwners(userAuth)
	if err!=nil{
		c.Status(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)
	}
	fmt.Printf("Sent Data : %+v",jsonData)
	c.JSON(http.StatusOK,jsonData)
}

func(o *OwnerController)AddOwner(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var owner models.OwnerReg
	err:=c.ShouldBindJSON(&owner)
	if err!=nil {
		fmt.Printf("error in json input:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=o.CreateOwner(userAuth.ID,&owner)
	if err!=nil{
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Owners created successfully",
	})
}

func(o *OwnerController)EditOwner(c *gin.Context){
	ownerID := c.Param("ownerID")
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var owner models.UserOutput
	owner.ID=ownerID
	err:=c.ShouldBindJSON(&owner)
	if err!=nil {
		fmt.Printf("error in parsing json: %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	if userAuth.Role==middleware.Admin {
		err = o.CheckOwnerCreator(userAuth.ID,owner.ID)
		if err != nil {
			if err!=database.ErrInternal {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
	}
	err=o.UpdateOwner(&owner)
	if err!=nil{
		fmt.Printf("err is %v",err)
		if err!=database.ErrInternal {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Owner updated successfully",
	})
}

func(o *OwnerController)DeleteOwners(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var ownerID struct {
		IDArr []string	`json:"idArr" binding:"required"`
	}
	err:=c.ShouldBindJSON(&ownerID)
	if err!=nil {
		fmt.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":ErrJsonInput,
		})
		return
	}
	err=o.RemoveOwners(userAuth,ownerID.IDArr...)
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
		"msg":"Owner deleted successfully",
	})
}
