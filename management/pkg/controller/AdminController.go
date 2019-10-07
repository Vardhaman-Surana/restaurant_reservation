package controller

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
)

type AdminController struct{
	database.Database
}

func NewAdminController(db database.Database)*AdminController{
	ac:=new(AdminController)
	ac.Database=db
	return ac
}

func(a *AdminController)GetAdmins(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"get_admins")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetAdmins",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	jsonData:=&[]models.UserOutput{}
	var stringData string
	var err error
	_,stringData,err=a.ShowAdmins(newCtx)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData!=""{
		_ = json.Unmarshal([]byte(stringData), jsonData)
	}
	w.WriteHeader(http.StatusOK)
	respBody,_:=json.Marshal(jsonData)
	w.Write(respBody)
}

func(a *AdminController)EditAdmin(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"update_admin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"EditAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	vars:=mux.Vars(rq)
	adminID:=vars["adminID"]
	var admin models.UserOutput
	admin.ID=adminID
	err:=json.NewDecoder(rq.Body).Decode(&admin)
	if err!=nil {
		models.WriteToResponse(w,http.StatusBadRequest,&models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	chkAdmCtx,err := a.CheckAdmin(newCtx,admin.ID)
	if err != nil {
		models.WriteToResponse(w,http.StatusBadRequest,&models.DefaultMap{
				"error":"Admin does not exist",
			})
			return
	}
	_,err=a.UpdateAdmin(chkAdmCtx,&admin)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"admin updated successfully",
	})
}

func(a *AdminController)DeleteAdmins(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"delete_admins")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteAdmins",ServiceName:tracing.ServiceName,}
	tracing.SetTags(span,tags)


	var adminID struct {
		IDArr []string	`json:"idArr" binding:"required"`
	}
	err:=json.NewDecoder(rq.Body).Decode(&adminID)
	if err!=nil {
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}
	_,err=a.RemoveAdmins(newCtx,adminID.IDArr...)
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
		"msg":"Admins deleted successfully",
	})
}