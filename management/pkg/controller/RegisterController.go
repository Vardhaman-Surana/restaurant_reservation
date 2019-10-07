package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
)

const ErrJsonInput="Invalid Json Input"

type RegisterController struct{
	database.Database
}

func NewRegisterController(db database.Database) *RegisterController{
	regController:=new(RegisterController)
	regController.Database=db
	return regController
}

func(r *RegisterController)Register(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"registration")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"Register",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var user models.UserReg
	err:=json.NewDecoder(rq.Body).Decode(&user)
	if err!=nil {
		fmt.Printf("err is %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
			"status":Fail,
		})
		return
	}
	if user.Role!=middleware.Admin && user.Role!=middleware.SuperAdmin {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_,err=r.CreateUser(newCtx,&user)
	if err!=nil{
		if err==database.ErrDupEmail{
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": database.ErrDupEmail.Error(),
				"status":Fail,
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"role":user.Role,
		"msg":"Registration Successful",
		"status":Success,
	})
}