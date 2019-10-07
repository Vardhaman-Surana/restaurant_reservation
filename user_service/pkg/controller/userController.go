package controller

import (
	"context"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/encryption"
	"github.com/vds/restaurant_reservation/user_service/pkg/fireBaseAuth"
	"github.com/vds/restaurant_reservation/user_service/pkg/jwtTokenGenerate"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
)

const (
	ErrInvalidJsonInput="invalid json input"
	ErrEmptyFields="missing required fields in input:"
	ErrInternal="internal server error"
	ErrInvalidEmail="email does not exist"
	ErrInCorrectPassword="incorrect password for the given email"
	ErrDupMail="email already used try with a different one"

	LogInSuccessfulMessage="LogIn Successful"
	RegistrationSuccessfulMessage="User Registration Successful"

	)
type UserController struct{
	db database.Database
}

func NewUserController(dbMap database.Database)*UserController{
	uc:=new(UserController)
	uc.db=dbMap
	return uc
}

func (uc *UserController)Register(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"user_registration")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"Register",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)


	var user models.User
	err:=json.NewDecoder(rq.Body).Decode(&user)
	if err!=nil {
		log.Printf("err is %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":nil,
			"error": ErrInvalidJsonInput,
		})
		return
	}
	errMsg:=ErrEmptyFields

	if user.Password==""{
		errMsg+=" password"
	}
	if user.Email==""{
		errMsg+=" email"
	}
	if user.Name==""{
		errMsg+=" name"
	}
	if errMsg!=ErrEmptyFields{
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":nil,
			"error": errMsg,
		})
		return
	}
	_,err=uc.db.CreateUser(newCtx,&user)
	if err!=nil{
		log.Printf("\nError in inserting user in Db : %v\n",err)
		if er, ok := err.(*mysql.MySQLError); ok {
			if er.Number == 1062 {
				models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
					"msg":nil,
					"error": ErrDupMail,
				})
				return
			}
		}
		models.WriteToResponse(w,http.StatusInternalServerError, &models.DefaultMap{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	models.WriteToResponse(w,http.StatusOK, &models.DefaultMap{
		"msg":RegistrationSuccessfulMessage,
		"error": nil,
	})
}

func (uc *UserController)LogIn(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"user_login")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"LogIn",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var user models.User

	err:=json.NewDecoder(rq.Body).Decode(&user)
	if err!=nil{
		log.Printf("err is %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":nil,
			"error": ErrInvalidJsonInput,
		})
		return
	}
	errMsg:=ErrEmptyFields
	if user.Password==""{
		errMsg+=" password"
	}
	if user.Email==""{
		errMsg+=" email"
	}
	if errMsg!=ErrEmptyFields{
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":nil,
			"error": errMsg,
		})
		return
	}
	getUserCtx,userOutput,err:=uc.db.GetUser(newCtx,user.Email)
	if err!=nil{
		log.Printf("err is %v",err)
		models.WriteToResponse(w,http.StatusBadRequest, &models.DefaultMap{
			"msg":nil,
			"error": ErrInvalidEmail,
		})
		return
	}
	matchPassCtx,isCorrect:=encryption.IsCorrectPassword(getUserCtx,userOutput.PasswordHash,user.Password)
	if !isCorrect{
		models.WriteToResponse(w,http.StatusUnauthorized, &models.DefaultMap{
			"msg":nil,
			"error": ErrInCorrectPassword,
		})
		return
	}
	/*claims:=map[string]interface{}{
		"ID":userOutput.ID,
	}*/
	/*
	_,token,err:=jwtTokenGenerate.CreateToken(matchPassCtx,claims)
	if err!=nil{
		log.Printf("error in generating token: %v",err)
		models.WriteToResponse(w,http.StatusInternalServerError, &models.DefaultMap{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}*/
	token,err:=fireBaseAuth.CreateToken(matchPassCtx,userOutput.ID)
	if err!=nil{
		log.Printf("error in generating token")
	}
	w.Header().Set("token",token)
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":LogInSuccessfulMessage,
		"error": nil,
	})
}


func(uc *UserController)LogOut(w http.ResponseWriter,rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx := tracing.GetSpanFromContext(prevCtx, "user_logout")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"LogOut",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	tokenStr:=rq.Header.Get("token")
	if tokenStr==""{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sTknCtx,err:=uc.db.StoreToken(newCtx,tokenStr)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go uc.db.DeleteExpiredToken(sTknCtx,tokenStr,jwtTokenGenerate.ExpireDuration)
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Logged Out Successfully",
		"error":nil,
	})
}
