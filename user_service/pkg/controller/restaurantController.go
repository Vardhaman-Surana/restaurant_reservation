package controller

import (
	"context"
	"encoding/json"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
)

type RestaurantController struct{
	db database.Database
}

func NewRestaurantController(db database.Database)*RestaurantController{
	resc:=new(RestaurantController)
	resc.db=db
	return resc
}

func(resc *RestaurantController)GetRestaurants(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx :=tracing.GetSpanFromContext(prevCtx,"get_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	_,restaurants,err:=resc.db.SelectRestaurants(newCtx)
	if err!=nil{
		log.Printf("error retrieving restaurants:%v",err)
		models.WriteToResponse(w,http.StatusInternalServerError,&models.DefaultMap{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	body,_:=json.Marshal(restaurants)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}