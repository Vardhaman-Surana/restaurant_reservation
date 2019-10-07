package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
	"strconv"
)

type MenuController struct{
	database.Database
}

func NewMenuController(db database.Database) *MenuController{
	menuController:=new(MenuController)
	menuController.Database=db
	return menuController
}

func (m *MenuController)GetMenu(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_restaurant_menu")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetMenu",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	res:=rq.Context().Value("restaurantID")
	resID:=res.(int)
	jsonData:=&[]models.DishOutput{}
	var stringData string
	_,stringData,err:=m.ShowMenu(newCtx,resID)
	if err!=nil{
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error":err.Error(),
		})
		return
	}
	if stringData!=""{
		_ =json.Unmarshal([]byte(stringData),jsonData)
	}
	w.WriteHeader(http.StatusOK)
	body,_:=json.Marshal(jsonData)
	w.Write(body)
}

func (m *MenuController)AddDishes(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"add_restaurant_dishes")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AddDishes",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	res:=rq.Context().Value("restaurantID")
	resID:=res.(int)
	var dishes []models.Dish
	err:=json.NewDecoder(rq.Body).Decode(&dishes)
	if err!=nil {
		fmt.Printf("error in reading json input:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=m.InsertDishes(newCtx,dishes,resID)
	if err!=nil{
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": err.Error(),
		})
		return
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg":"Dishes Added to menu successfully",
	})
	return
}

func (m *MenuController)EditDish(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"update_restaurant_menu")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"EditDish",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var dish models.DishOutput
	res:=rq.Context().Value("restaurantID")
	resID:=res.(int)
	vars:=mux.Vars(rq)
	dishValue:=vars["dishID"]
	dishID,_:=strconv.Atoi(dishValue)
	dish.ID=dishID
	err:=json.NewDecoder(rq.Body).Decode(&dish)
	if err!=nil {
		fmt.Printf("error in parsing json input:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	chkDishCtx,err:=m.CheckRestaurantDish(newCtx,resID,dish.ID)
	if err!=nil{
		if err!=database.ErrInternal{
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_,err=m.UpdateDish(chkDishCtx,&dish)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg":"Dish Updated successfully",
	})
}

func(m *MenuController)DeleteDishes(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"delete_restaurant_dishes")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteDishes",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var dishID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err:=json.NewDecoder(rq.Body).Decode(&dishID)
	if err!=nil {
		fmt.Printf("error in reading json input:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=m.RemoveDishes(newCtx,dishID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": err.Error(),
			})
			return
		}
	}
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Dishes deleted successfully",
	})
}