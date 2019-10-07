package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"log"
	"net/http"
)

const(
	Success="Success"
	Fail="Fail"
)

type LogInController struct{
	database.Database
}

func NewLogInController(db database.Database)*LogInController{
	lc:=new(LogInController)
	lc.Database=db
	return lc
}
func(l *LogInController)LogIn(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"login")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"LogIn",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var cred models.Credentials
	err:=json.NewDecoder(rq.Body).Decode(&cred)
	if err!=nil {
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"error": ErrJsonInput,
		})
		return
	}

	isValid:=middleware.IsValidUserType(cred.Role)
	if !isValid{
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logInCtx,userID,err:=l.LogInUser(newCtx,&cred)
	if err!=nil{
		fmt.Println(err)
		models.WriteToResponse(w,http.StatusUnauthorized, &models.DefaultMap{
			"error": database.ErrInvalidCredentials.Error(),
			"status": Fail,
		})
		return
	}
	claims:=&models.Claims{
		ID:userID,
		Role:cred.Role,
	}
	_,token,err:=encryption.CreateToken(logInCtx,claims)
	if err!=nil{
		log.Printf("%v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("token",token)
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"role":cred.Role,
		"msg":"Login Successful",
		"status":Success,
	})
}

func(l *LogInController)LogOut(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"logout")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"LogOut",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	tokenStr:=rq.Header.Get("token")
	if tokenStr==""{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	storeTokenCtx,err:=l.StoreToken(newCtx,tokenStr)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go l.DeleteExpiredToken(storeTokenCtx,tokenStr,encryption.ExpireDuration)
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Logged Out Successfully",
		"status":Success,
	})
}


