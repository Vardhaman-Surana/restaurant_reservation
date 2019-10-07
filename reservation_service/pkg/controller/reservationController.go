package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/middleware"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/models"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/tracing"
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
	rc:=new(ReservationController)
	rc.db=db
	return rc
}

func (rc *ReservationController)CheckAvailability(w http.ResponseWriter,rq *http.Request) {
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx :=tracing.GetSpanFromContext(prevCtx,"check_reservation_availability")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckAvailability",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	resIDString:= rq.URL.Query().Get(resIDParam)
	errMsg := ErrQueryParamNotFound
	if resIDString=="" {
		errMsg = errMsg + " " + resIDParam
	}
	startTimeString:= rq.URL.Query().Get(startTimeParam)
	if startTimeString=="" {
		errMsg = errMsg + " " + startTimeParam
	}
	if errMsg != ErrQueryParamNotFound {
		models.WriteToResponse(w,http.StatusBadRequest,&models.DefaultMap{
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
		models.WriteToResponse(w,http.StatusBadRequest,&models.DefaultMap{
			"msg":   nil,
			"error": errMsg,
		})
		return
	}
	if int64(startTime+OneMinute) < time.Now().Unix(){
		models.WriteToResponse(w,http.StatusNotAcceptable,&models.DefaultMap{
			"msg":   nil,
			"error": "entered startTime is of the past",
		})
		return
	}
	_,numTables,err:=rc.db.GetNumAvailableTables(newCtx,resID,int64(startTime))
	if err!=nil{
		log.Printf("\nerror in getting the number of tables:%v\n",err)
		models.WriteToResponse(w,http.StatusInternalServerError,&models.DefaultMap{
			"msg":   nil,
			"error": ErrInternal,
		})
		return
	}
	if numTables==0{
		models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
			"msg":   ReservationNotAvailableMessage,
			"error": nil,
		})
		return
	}
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":   ReservationAvailableMessage+fmt.Sprintf("%d",numTables),
		"error": nil,
	})
}

func (rc *ReservationController)AddReservation(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx :=tracing.GetSpanFromContext(prevCtx,"make_reservation")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AddReservation",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	userID:=rq.Context().Value(middleware.UserIDContextKey)
	data:=struct{
		ResID int `json:"resID"`
		StartTime int64 `json:"startTime"`
	}{}
	err:=json.NewDecoder(rq.Body).Decode(&data)
	if err!=nil{
		log.Printf("err is %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
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
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":   nil,
			"error": errMsg,
		})
		return
	}
	if (data.StartTime+OneMinute) < time.Now().Unix(){
		models.WriteToResponse(w,http.StatusNotAcceptable, &models.DefaultMap{
			"msg":   nil,
			"error": "entered startTime is of the past",
		})
		return
	}

	_,id,err:=rc.db.CreateReservation(newCtx,data.ResID,data.StartTime,userID.(string))
	if err!=nil{
		log.Printf("error in creating reservations:%v",err)
		if err.Error()==ReservationNotAvailableMessage{
			models.WriteToResponse(w,http.StatusNotAcceptable, &models.DefaultMap{
				"msg":   nil,
				"error": ReservationNotAvailableMessage,
			})
			return
		}
		models.WriteToResponse(w,http.StatusInternalServerError, &models.DefaultMap{
			"msg":   nil,
			"error": ErrInternal,
		})
		return
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg":   ReservationSuccessMessage,
		"error": nil,
		"resvID":id,
	})
}
