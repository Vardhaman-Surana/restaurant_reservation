package controller

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/queue"
	"log"
	"net/http"
	"time"
)

type RestaurantController struct{
	database.Database
	Logger *fluent.Fluent
}

func NewRestaurantController(db database.Database,logger *fluent.Fluent) *RestaurantController{
	resController:=new(RestaurantController)
	resController.Database=db
	resController.Logger=logger
	return resController
}

func(r *RestaurantController)GetNearBy(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	var location models.Location
	err:=c.ShouldBindJSON(&location)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	var jsonData=[]struct{
		Name string `json:"name" binding:"required"`
	}{}
	stringData,err:=r.ShowNearBy(&location)
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),&jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func(r *RestaurantController)GetRestaurants(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	var stringData string
	var err error
	stringData, err = r.ShowRestaurants(userAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (r *RestaurantController)AddRestaurant(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var restaurant models.Restaurant
	err:=c.ShouldBindJSON(&restaurant)
	if err!=nil {
		fmt.Printf("error parsing json input")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	restaurant.CreatorID=userAuth.ID
	resId,err:=r.InsertRestaurant(&restaurant)
	if err!=nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if numTables:=restaurant.NumTables;numTables!=0{
		err:=rabbitmq_queue.SendMessage(resId,numTables,r.Logger)
		if err!=nil{
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK,gin.H{
		"msg":"Restaurant added",
	})
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (r *RestaurantController)EditRestaurant(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=middleware.Admin && userAuth.Role!=middleware.SuperAdmin{
		c.Status(http.StatusUnauthorized)
		return
	}
	res,_:=c.Get("restaurantID")
	resID:=res.(int)
	var restaurant models.RestaurantOutput
	restaurant.ID=resID
	err:=c.ShouldBindJSON(&restaurant)
	if err!=nil {
		fmt.Printf("error in parsing json:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=r.UpdateRestaurant(&restaurant)
	if err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Restaurant Updated Successfully",
	})
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (r *RestaurantController)DeleteRestaurants(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err:=c.ShouldBindJSON(&resID)
	if err!=nil {
		fmt.Printf("error in parsing json input:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=r.RemoveRestaurants(userAuth,resID.IDArr...)
	if err!=nil{
		fmt.Printf("err is %v",err)
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Restaurants deleted Successfully",
	})
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}
func(r *RestaurantController)GetOwnerRestaurants(c * gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	ownerID := c.Param("ownerID")
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	if userAuth.Role==middleware.Admin {
		err = r.CheckOwnerCreator(userAuth.ID,ownerID)
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
	jsonData:=&[]models.RestaurantOutput{}
	ownerAuth:=models.UserAuth{
		ID:   ownerID,
		Role: middleware.Owner,
	}
	var stringData string
	stringData, err = r.ShowRestaurants(&ownerAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)
	}
	c.JSON(http.StatusOK,jsonData)
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (r *RestaurantController)GetAvailableRestaurants(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	stringData,err:=r.ShowAvailableRestaurants(userAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)

	}
	c.JSON(http.StatusOK,jsonData)
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}

func (r *RestaurantController)AddOwnerForRestaurants(c *gin.Context){
	er:=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Serving Request")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	ownerID := c.Param("ownerID")
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	if userAuth.Role==middleware.Admin {
		err = r.CheckOwnerCreator(userAuth.ID,ownerID)
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
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err=c.ShouldBindJSON(&resID)
	if err!=nil {
		fmt.Printf("error in parsing json:%v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrJsonInput,
		})
		return
	}
	err=r.InsertOwnerForRestaurants(userAuth,ownerID,resID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Owner assigned restaurants Successfully",
	})
	er=r.Logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf("%v",c.Request.URL),"info":fmt.Sprintf("Served")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
}