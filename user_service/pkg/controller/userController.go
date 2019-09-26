package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/encryption"
	"github.com/vds/restaurant_reservation/user_service/pkg/jwtTokenGenerate"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"net/http"
	"time"
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
	tracer opentracing.Tracer
}

func NewUserController(dbMap database.Database,	tracer opentracing.Tracer)*UserController{
	uc:=new(UserController)
	uc.db=dbMap
	uc.tracer=tracer
	return uc
}

func (uc *UserController)Register(c *gin.Context){
	opentracing.SetGlobalTracer(uc.tracer)
	span, newCtx := opentracing.StartSpanFromContext(c, "user_register")
	span.SetBaggageItem("requestID", uuid.New().String())
	span.SetBaggageItem("requestUrl",c.Request.URL.String())
	span.SetTag("funcName","Register")
	span.SetTag("serviceName",tracing.ServiceName)
	span.SetTag("startTime",time.Now().String())
	defer span.Finish()


	var user models.User
	err:=c.ShouldBindJSON(&user)
	if err!=nil {
		log.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":nil,
			"error": errMsg,
		})
		return
	}
	err=uc.db.InsertUser(newCtx,&user)
	if err!=nil{
		log.Printf("\nError in inserting user in Db : %v\n",err)
		if er, ok := err.(*mysql.MySQLError); ok {
			if er.Number == 1062 {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg":nil,
					"error": ErrDupMail,
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":RegistrationSuccessfulMessage,
		"error": nil,
	})
}

func (uc *UserController)LogIn(c *gin.Context){
	opentracing.SetGlobalTracer(uc.tracer)
	span, newCtx := opentracing.StartSpanFromContext(c, "user_login")
	span.SetBaggageItem("requestID", uuid.New().String())
	span.SetBaggageItem("requestUrl",c.Request.URL.String())
	span.SetTag("funcName","LogIn")
	span.SetTag("serviceName",tracing.ServiceName)
	span.SetTag("startTime",time.Now().String())
	defer span.Finish()

	var user models.User

	err:=c.ShouldBindJSON(&user)
	if err!=nil{
		log.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":nil,
			"error": errMsg,
		})
		return
	}
	userOutput,err:=uc.db.GetUser(newCtx,user.Email)
	if err!=nil{
		log.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":nil,
			"error": ErrInvalidEmail,
		})
		return
	}
	if !encryption.IsCorrectPassword(userOutput.PasswordHash,user.Password){
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg":nil,
			"error": ErrInCorrectPassword,
		})
		return
	}

	claims:=&models.Claims{
		ID:userOutput.ID,
	}
	token,err:=jwtTokenGenerate.CreateToken(claims)
	if err!=nil{
		log.Printf("error in generating token: %v",err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":nil,
			"error": ErrInternal,
		})
		return
	}
	c.Writer.Header().Set("token",token)
	c.JSON(http.StatusOK,gin.H{
		"msg":LogInSuccessfulMessage,
		"error": nil,
	})
}


func(uc *UserController)LogOut(c *gin.Context){
	prevContext,_:=c.Get("context")
	prevCtx:=prevContext.(context.Context)
	span, newCtx := opentracing.StartSpanFromContext(prevCtx, "user_logout")
	span.SetBaggageItem("requestID", uuid.New().String())
	span.SetBaggageItem("requestUrl",c.Request.URL.String())
	span.SetTag("serviceName",tracing.ServiceName)
	span.SetTag("funcName","LogOut")
	span.SetTag("startTime",time.Now().String())
	defer span.Finish()

	tokenStr:=c.Request.Header.Get("token")
	if tokenStr==""{
		c.Status(http.StatusBadRequest)
		return
	}
	err:=uc.db.StoreToken(newCtx,tokenStr)
	if err!=nil{
		c.Status(http.StatusInternalServerError)
		return
	}
	go uc.db.DeleteExpiredToken(newCtx,tokenStr,jwtTokenGenerate.ExpireDuration)
	c.JSON(http.StatusOK,gin.H{
		"msg":"Logged Out Successfully",
		"error":nil,
	})
}
