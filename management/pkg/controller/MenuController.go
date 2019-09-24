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
	"strconv"
	"time"
)

type MenuController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewMenuController(db database.Database,logger *fluent.Fluent) *MenuController{
	menuController:=new(MenuController)
	menuController.Database=db
	menuController.Logger=logger
	return menuController
}

func (m *MenuController)GetMenu(c *gin.Context){
	er:=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	res,_:=c.Get("restaurantID")
	resID:=res.(int)
	jsonData:=&[]models.DishOutput{}
	var stringData string
	stringData,err:=m.ShowMenu(resID)
	if err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error":err.Error(),
		})
		return
	}
	if stringData!=""{
		_ =json.Unmarshal([]byte(stringData),jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
	er=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (m *MenuController)AddDishes(c *gin.Context){
	er:=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	res,_:=c.Get("restaurantID")
	resID:=res.(int)
	var dishes []models.Dish
	err:=c.ShouldBind(&dishes)
	if err!=nil {
		fmt.Printf("error in reading json input:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=m.InsertDishes(dishes,resID)
	if err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":"Dishes Added to menu successfully",
	})
	er=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (m *MenuController)EditDish(c *gin.Context){
	er:=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	var dish models.DishOutput
	res,_:=c.Get("restaurantID")
	resID:=res.(int)
	dishValue:=c.Param("dishID")
	dishID,_:=strconv.Atoi(dishValue)
	dish.ID=dishID
	err:=c.ShouldBindJSON(&dish)
	if err!=nil {
		fmt.Printf("error in parsing json input:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=m.CheckRestaurantDish(resID,dish.ID)
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
	err=m.UpdateDish(&dish)
	if err!=nil{
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":"Dish Updated successfully",
	})
	er=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func(m *MenuController)DeleteDishes(c *gin.Context){
	er:=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	var dishID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err:=c.ShouldBindJSON(&dishID)
	if err!=nil {
		fmt.Printf("error in reading json input:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=m.RemoveDishes(dishID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Dishes deleted successfully",
	})
	er=m.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}