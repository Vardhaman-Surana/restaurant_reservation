package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/queue"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
)

type RestaurantController struct{
	database.Database
}

func NewRestaurantController(db database.Database) *RestaurantController{
	resController:=new(RestaurantController)
	resController.Database=db
	return resController
}

func(r *RestaurantController)GetNearBy(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_nearby_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetNearBy",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	var location models.Location
	err:=json.NewDecoder(rq.Body).Decode(&location)
	if err!=nil {
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	var jsonData=[]struct{
		Name string `json:"name" binding:"required"`
	}{}
	_,stringData,err:=r.ShowNearBy(newCtx,&location)
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),&jsonData)
	}
	body,_:=json.Marshal(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func(r *RestaurantController)GetRestaurants(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	var stringData string
	var err error
	_,stringData, err = r.ShowRestaurants(newCtx,userAuth)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)
	}
	body,_:=json.Marshal(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Write(body)}

func (r *RestaurantController)AddRestaurant(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"add_restaurant")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AddRestaurant",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var restaurant models.Restaurant
	err:=json.NewDecoder(rq.Body).Decode(&restaurant)
	if err!=nil {
		fmt.Printf("error parsing json input")
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	restaurant.CreatorID=userAuth.ID
	insertCtx,resId,err:=r.InsertRestaurant(newCtx,&restaurant)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if numTables:=restaurant.NumTables;numTables!=0{
		_,err:=rabbitmq_queue.SendMessage(insertCtx,resId,numTables,)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Restaurant added",
	})
}

func (r *RestaurantController)EditRestaurant(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"update_restaurant")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"EditRestaurant",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=middleware.Admin && userAuth.Role!=middleware.SuperAdmin{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res:=rq.Context().Value("restaurantID")
	resID:=res.(int)
	var restaurant models.RestaurantOutput
	restaurant.ID=resID
	err:=json.NewDecoder(rq.Body).Decode(&restaurant)
	if err!=nil {
		fmt.Printf("error in parsing json:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=r.UpdateRestaurant(newCtx,&restaurant)
	if err!=nil{
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": err.Error(),
		})
		return
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg": "Restaurant Updated Successfully",
	})
}

func (r *RestaurantController)DeleteRestaurants(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"delete_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err:=json.NewDecoder(rq.Body).Decode(&resID)
	if err!=nil {
		fmt.Printf("error in parsing json input:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=r.RemoveRestaurants(newCtx,userAuth,resID.IDArr...)
	if err!=nil{
		fmt.Printf("err is %v",err)
		if err!=database.ErrInternal{
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": err.Error(),
			})
			return
		}
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg": "Restaurants deleted Successfully",
	})
}
func(r *RestaurantController)GetOwnerRestaurants(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_restaurant_of_an_owner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetOwnerRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	vars:=mux.Vars(rq)
	ownerID:=vars["ownerID"]
	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	var chkOwnCtr context.Context
	if userAuth.Role==middleware.Admin {
		chkOwnCtr,err = r.CheckOwnerCreator(newCtx,userAuth.ID,ownerID)
		if err != nil {
			if err!=database.ErrInternal {
				models.WriteToResponse(w,http.StatusUnauthorized, &models.DefaultMap{
					"error": err.Error(),
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	jsonData:=&[]models.RestaurantOutput{}
	ownerAuth:=models.UserAuth{
		ID:   ownerID,
		Role: middleware.Owner,
	}
	var stringData string
	_,stringData, err = r.ShowRestaurants(chkOwnCtr,&ownerAuth)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)
	}
	body,_:=json.Marshal(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Write(body)}

func (r *RestaurantController)GetAvailableRestaurants(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_aval_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetAvailableRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	_,stringData,err:=r.ShowAvailableRestaurants(newCtx,userAuth)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_=json.Unmarshal([]byte(stringData),jsonData)

	}
	body,_:=json.Marshal(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (r *RestaurantController)AddOwnerForRestaurants(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"assign_owner_to_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AddOwnerForRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	vars:=mux.Vars(rq)
	ownerID:=vars["ownerID"]
	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	var chkOwnCtr context.Context

	if userAuth.Role==middleware.Admin {
		chkOwnCtr,err = r.CheckOwnerCreator(newCtx,userAuth.ID,ownerID)
		if err != nil {
			if err!=database.ErrInternal {
				models.WriteToResponse(w,http.StatusUnauthorized, &models.DefaultMap{
					"error": err.Error(),
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err=json.NewDecoder(rq.Body).Decode(&resID)
	if err!=nil {
		fmt.Printf("error in parsing json:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=r.InsertOwnerForRestaurants(chkOwnCtr,userAuth,ownerID,resID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": err.Error(),
			})
			return
		}
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg": "Owner assigned restaurants Successfully",
	})
}