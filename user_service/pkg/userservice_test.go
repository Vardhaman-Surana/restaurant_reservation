package user_service_test

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/server"
	"github.com/vds/restaurant_reservation/user_service/pkg/testHelpers"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)
const (
	ExpiredToken ="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImQyZWExMDcxLTMzMjYtNDI1Mi05ODQzLWQ5YWJiMWQ1NDI5MSIsImV4cCI6MTU2ODgwNjY3Nn0.zUmVKIDzOA6YBx5pXMEeBxGOYi74panyuf_mZCiwhBo"

)
func TestUserService(t *testing.T){
	DB,err:=testHelpers.InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	r,err:=server.NewRouter(DB)
	if err!=nil{
		t.Fatalf("could not get server instance")
	}
	router:=r.Create()
	t.Run("test for invalid json input when register",func(t *testing.T){
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

	firstDataRegister:=models.User{Email:"",Password:"",Name:""}
	firstResponseRegister:=map[string]interface{}{"msg":nil,"error":controller.ErrEmptyFields+" password email name"}

	secondDataRegister:=models.User{Email:"abc@gmail.com",Password:"abc",Name:"abc"}
	secondResponseRegister:=map[string]interface{}{"msg":controller.RegistrationSuccessfulMessage,"error":nil}

	thirdDataRegister:=models.User{Email:"abc@gmail.com",Password:"abc",Name:"abc"}
	thirdResponseRegister:=map[string]interface{}{"msg":nil,"error":controller.ErrDupMail}

	testsRegister:= []struct{
		name string
		data *models.User
		wantStatus int
		wantResponse map[string]interface{}
	}{
		{"registration with empty fields",&firstDataRegister,http.StatusBadRequest,firstResponseRegister},
		{"successful registration",&secondDataRegister,http.StatusOK,secondResponseRegister},
		{"registration with a duplicate email",&thirdDataRegister,http.StatusBadRequest,thirdResponseRegister},
	}
	for _,test :=range testsRegister{
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

	t.Run("test for invalid json input for login",func(t *testing.T){
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

	firstDataLogin:=models.User{Email:"",Password:""}
	firstResponseLogin:=map[string]interface{}{"msg":nil,"error":controller.ErrEmptyFields+" password email"}

	secondResponseLogin:=map[string]interface{}{"msg":controller.LogInSuccessfulMessage,"error":nil}

	thirdDataLogin:=models.User{Email:"vardhaman@gmail.co",Password:"passwor"}
	thirdResponseLogin:=map[string]interface{}{"msg":nil,"error":controller.ErrInvalidEmail}

	fourthDataLogin:=models.User{Email:"vardhaman@gmail.com",Password:"passwor"}
	fourthResponseLogin:=map[string]interface{}{"msg":nil,"error":controller.ErrInCorrectPassword}

	var userToken string
	testsLogin:= []struct{
		name string
		data *models.User
		wantStatus int
		wantResponse map[string]interface{}
	}{
		{"login with empty fields",&firstDataLogin,http.StatusBadRequest,firstResponseLogin},
		{"login successful",&secondDataRegister,http.StatusOK,secondResponseLogin},
		{"invalid email",&thirdDataLogin,http.StatusBadRequest,thirdResponseLogin},
		{"incorrect password",&fourthDataLogin,http.StatusUnauthorized,fourthResponseLogin},
	}
	for _,test :=range testsLogin{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewLogInRequest(test.data)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			testHelpers.AssertResponse(t, response.Body.String(), test.wantResponse)
			if response.Code==200{
				userToken=response.Header().Get("token")
			}
		})
	}

	respGetRes1:=[]byte(`[{"id":3,"name":"dummyRestaurant","lat":0.1,"lng":0.2},{"id":4,"name":"dummyRestaurant1","lat":1,"lng":2}]`)

	respGetRes2:=map[string]interface{}{"msg":nil,"error":middleware.TokenExpireErr}


	testGetRestaurants:=[]struct{
		name string
		token string
		wantStatus int
		resp interface{}
	}{
		{"get restaurants with a valid token",userToken,http.StatusOK,respGetRes1},
		{"get restaurants with invalid token","invalid",http.StatusUnauthorized,nil},
		{"get restaurants with expired token",ExpiredToken,http.StatusUnauthorized,respGetRes2},
	}
	for _,test:=range testGetRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewGetRestaurantRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			if response.Code==200 {
				testHelpers.AssertGetResponse(t, response.Body.Bytes(), test.resp.([]byte))
			}else if test.resp!=nil{
				testHelpers.AssertResponse(t,response.Body.String(),test.resp.(map[string]interface{}))
			}
		})
	}


}

