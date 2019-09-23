package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/middleware"
	"log"
	"net/http"
	"strconv"
	"time"
)

const(
	ErrInternal = "internal server error"
	ErrQueryParamNotFound = "required query parameters missing:"
	ErrInvalidParamType = "some required query parameters are not of type integer:"
	ErrInvalidJsonInput = "invaild json input"
	resIDParam = "resID"
	startTimeParam = "startTime"
	ErrEmptyFields ="some required fields missing:"

	ReservationNotAvailableMessage = "reservations not available"
	ReservationAvailableMessage = "tables available for reservation : "
	ReservationSuccessMessage = "reservation successful"

	OneMinute=60//unix seconds
)


type ReservationController struct{
	db database.Database
}

func NewReservationController(db database.Database)*ReservationController{
	resc:=new(ReservationController)
	resc.db=db
	return resc
}

func (rc *ReservationController)CheckAvailability(c *gin.Context) {
	resIDString, ok := c.GetQuery(resIDParam)
	errMsg := ErrQueryParamNotFound
	if !ok {
		errMsg = errMsg + " " + resIDParam
	}
	startTimeString, ok := c.GetQuery(startTimeParam)
	if !ok {
		errMsg = errMsg + " " + startTimeParam
	}
	if errMsg != ErrQueryParamNotFound {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":   nil,
			"error": errMsg,
		})
		return
	}
	errMsg = ErrInvalidParamType
	startTime, err := strconv.Atoi(startTimeString)
	if err != nil {
		errMsg = errMsg+" "+ startTimeParam
	}
	resID, err := strconv.Atoi(resIDString)
	if err != nil {
		errMsg = errMsg+" "+ resIDParam
	}
	if errMsg!= ErrInvalidParamType{
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":   nil,
			"error": errMsg,
		})
		return
	}
	if int64(startTime+OneMinute) < time.Now().Unix(){
		c.JSON(http.StatusNotAcceptable, gin.H{
			"msg":   nil,
			"error": "entered startTime is of the past",
		})
		return
	}
	numTables,err:=rc.db.GetNumAvailableTables(resID,int64(startTime))
	if err!=nil{
		log.Printf("\nerror in getting the number of tables:%v\n",err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   nil,
			"error": ErrInternal,
		})
		return
	}
	if numTables==0{
		c.JSON(http.StatusOK, gin.H{
			"msg":   ReservationNotAvailableMessage,
			"error": nil,
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":   ReservationAvailableMessage+fmt.Sprintf("%d",numTables),
		"error": nil,
	})
}

func (rc *ReservationController)AddReservation(c *gin.Context){
	userID,exists:=c.Get(middleware.UserIDContextKey)
	if !exists{
		log.Printf("\nDid not get user id in the context got: <%v> instead\n",userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   nil,
			"error": ErrInternal,
		})
	}
	data:=struct{
		ResID int `json:"resID"`
		StartTime int64 `json:"startTime"`
	}{}
	err:=c.ShouldBindJSON(&data)
	if err!=nil{
		log.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":nil,
			"error": ErrInvalidJsonInput,
		})
		return
	}
	errMsg := ErrEmptyFields
	if data.ResID==0 {
		errMsg = errMsg +" "+resIDParam
	}
	if data.StartTime==0 {
		errMsg = errMsg + " " + startTimeParam
	}
	if errMsg != ErrEmptyFields {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":   nil,
			"error": errMsg,
		})
		return
	}
	if (data.StartTime+OneMinute) < time.Now().Unix(){
		c.JSON(http.StatusNotAcceptable, gin.H{
			"msg":   nil,
			"error": "entered startTime is of the past",
		})
		return
	}

	id,err:=rc.db.CreateReservation(data.ResID,data.StartTime,userID.(string))
	if err!=nil{
		log.Printf("error in creating reservations:%v",err)
		if err.Error()==ReservationNotAvailableMessage{
			c.JSON(http.StatusNotAcceptable, gin.H{
				"msg":   nil,
				"error": ReservationNotAvailableMessage,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":   nil,
			"error": ErrInternal,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   ReservationSuccessMessage,
		"error": nil,
		"resvID":id,
	})
}
