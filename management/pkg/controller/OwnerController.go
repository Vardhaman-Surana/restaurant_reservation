package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
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
func(o *OwnerController)GetOwners(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_owners")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetOwners",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.UserOutput{}
	var stringData string
	var err error
	_,stringData,err=o.ShowOwners(newCtx,userAuth)
	if err!=nil{
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

func(o *OwnerController)AddOwner(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"add_owner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"AddOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var owner models.OwnerReg
	err:=json.NewDecoder(rq.Body).Decode(&owner)
	if err!=nil {
		fmt.Printf("error in json input:%v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=o.CreateOwner(newCtx,userAuth.ID,&owner)
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
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Owners created successfully",
	})
}

func(o *OwnerController)EditOwner(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"update_owner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"EditOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	vars:=mux.Vars(rq)
	ownerID:=vars["ownerID"]
	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var owner models.UserOutput
	owner.ID=ownerID
	err:=json.NewDecoder(rq.Body).Decode(&owner)
	if err!=nil {
		fmt.Printf("error in parsing json: %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	var chkCtrCtx context.Context
	if userAuth.Role==middleware.Admin {
		chkCtrCtx,err = o.CheckOwnerCreator(newCtx,userAuth.ID,owner.ID)
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
	_,err=o.UpdateOwner(chkCtrCtx,&owner)
	if err!=nil{
		fmt.Printf("err is %v",err)
		if err!=database.ErrInternal {
			models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
				"error": err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Owner updated successfully",
	})
}

func(o *OwnerController)DeleteOwners(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"delete_owners")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteOwners",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	var ownerID struct {
		IDArr []string	`json:"idArr" binding:"required"`
	}
	err:=json.NewDecoder(rq.Body).Decode(&ownerID)
	if err!=nil {
		fmt.Print(err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error":ErrJsonInput,
		})
		return
	}
	_,err=o.RemoveOwners(newCtx,userAuth,ownerID.IDArr...)
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
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Owner deleted successfully",
	})
}
