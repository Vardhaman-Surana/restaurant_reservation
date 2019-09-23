package controller_test

import (
	"fmt"
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"

	"github.com/vds/restaurant_reservation/user_service/pkg/testHelpers"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestRegistration(t *testing.T){
	DB,err:=testHelpers.InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer testHelpers.CleanDB(DB)
	router,err:=testHelpers.GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}
	firstData:=models.User{Email:"",Password:"",Name:""}
	firstResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrEmptyFields+" password email name"}

	secondData:=models.User{Email:"abc@gmail.com",Password:"abc",Name:"abc"}
	secondResponse:=map[string]interface{}{"msg":controller.RegistrationSuccessfulMessage,"error":nil}

	thirdData:=models.User{Email:"abc@gmail.com",Password:"abc",Name:"abc"}
	thirdResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrDupMail}

	tests:= []struct{
		name string
		data *models.User
		wantStatus int
		wantResponse map[string]interface{}
	}{
		{"registration with empty fields",&firstData,http.StatusBadRequest,firstResponse},
		{"successful registration",&secondData,http.StatusOK,secondResponse},
		{"registration with a duplicate email",&thirdData,http.StatusBadRequest,thirdResponse},
	}
	for _,test :=range tests{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewRegisterRequest(test.data)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			if resp:=response.Body.String();resp!="" {
				testHelpers.AssertResponse(t, response.Body.String(), test.wantResponse)
			}
			if response.Code==200{
				testHelpers.AssertUserEntry(t,DB,test.data)
			}
		})
	}
	t.Run("test for invalid json input",func(t *testing.T){
		data:=`{"email":"hi",}`
		req,err:=http.NewRequest(http.MethodPost, "/register", strings.NewReader(data))
		if err!=nil{
			log.Fatal("error forming a registration request")
		}
		req.Header.Add("Content-Type", "application/json")
		response:=httptest.NewRecorder()
		router.ServeHTTP(response,req)
		testHelpers.AssertStatus(t,response.Code,http.StatusBadRequest)
		wantResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrInvalidJsonInput}
		testHelpers.AssertResponse(t,response.Body.String(),wantResponse)
	})
}

func TestLogIn(t *testing.T){
	DB,err:=testHelpers.InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer testHelpers.CleanDB(DB)
	router,err:=testHelpers.GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}
	firstData:=models.User{Email:"",Password:""}
	firstResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrEmptyFields+" password email"}

	secondData:=models.User{Email:"vardhaman@gmail.com",Password:"password"}
	secondResponse:=map[string]interface{}{"msg":controller.LogInSuccessfulMessage,"error":nil}

	thirdData:=models.User{Email:"vardhaman@gmail.co",Password:"passwor"}
	thirdResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrInvalidEmail}

	fourthData:=models.User{Email:"vardhaman@gmail.com",Password:"passwor"}
	fourthResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrInCorrectPassword}

	tests:= []struct{
		name string
		data *models.User
		wantStatus int
		wantResponse map[string]interface{}
	}{
		{"login with empty fields",&firstData,http.StatusBadRequest,firstResponse},
		{"login successful",&secondData,http.StatusOK,secondResponse},
		{"invalid email",&thirdData,http.StatusBadRequest,thirdResponse},
		{"incorrect password",&fourthData,http.StatusUnauthorized,fourthResponse},
	}
	for _,test :=range tests{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewLogInRequest(test.data)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			if resp:=response.Body.String();resp!="" {
				testHelpers.AssertResponse(t, response.Body.String(), test.wantResponse)
			}
		})
	}
	t.Run("test for invalid json input",func(t *testing.T){
		data:=`{"email":"hi",}`
		req,err:=http.NewRequest(http.MethodPost, "/login", strings.NewReader(data))
		if err!=nil{
			log.Fatal("error forming a registration request")
		}
		req.Header.Add("Content-Type", "application/json")
		response:=httptest.NewRecorder()
		router.ServeHTTP(response,req)
		testHelpers.AssertStatus(t,response.Code,http.StatusBadRequest)
		wantResponse:=map[string]interface{}{"msg":nil,"error":controller.ErrInvalidJsonInput}
		testHelpers.AssertResponse(t,response.Body.String(),wantResponse)
	})
}

func TestLogOut(t *testing.T){
	DB,err:=testHelpers.InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer testHelpers.CleanDB(DB)
	router,err:=testHelpers.GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}

	token:=testHelpers.GetUserToken(router)

	resp1:=map[string]interface{}{"msg":"Logged Out Successfully","error":nil}

	///tests for logout
	testLogout:=[]struct{
		name string
		token string
		wantStatus int
		wantResp map[string]interface{}
		tableName string
	}{
		{"when request is made with token",token,http.StatusOK,resp1,testHelpers.TokenTable},
		{"when token is not sent","",http.StatusBadRequest,nil,""},
	}
	for _,test:=range testLogout{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewLogOutRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			if response.Code==200 {
				testHelpers.AssertResponse(t, response.Body.String(), test.wantResp)
				assertTokenEntry(t, DB, test.token, test.tableName)
			}
		})
	}
}
func assertTokenEntry(t *testing.T,db *mysql.MysqlDbMap,wantToken ,tableName string){
	if tableName==""{
		return
	}
	gotToken:=""
	err:=db.DbMap.SelectOne(&gotToken,fmt.Sprintf("select token from %s where token=?",tableName),wantToken)
	if err!=nil{
		log.Fatalf("can not perform query:%v",err)
	}
	if gotToken==""{
		t.Error("token  not  stored in database")
	}

}

//functions used in tests
