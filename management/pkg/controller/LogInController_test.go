package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	_ "testing"
)
var (
	dummyAdmin=models.UserReg{"admin","dummyAdmin@gmail.com","dummyAdmin","dummyPass"}
 	dummySuperAdmin=models.UserReg{"superAdmin","dummySuperAdmin@gmail.com","dummySuperAdmin","dummySuperPass"}
 	dummyOwner=models.OwnerReg{"dummySuperOwner@gmail.com","dummySuperOwner","dummyOwnerPass"}
)
const TokenTable="invalid_tokens"

func TestLogInController(t *testing.T){
	DB,err:=InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer CleanDB(DB)
	router,err:=GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}
	//data for valid credentials
	first:=&models.Credentials{dummyAdmin.Role,dummyAdmin.Email,dummyAdmin.Password}
	resp1:=map[string]string{"role":first.Role,"msg":"Login Successful","status":controller.Success}

	//data for invalid credentials
	second:=&models.Credentials{dummySuperAdmin.Role,dummySuperAdmin.Email,"dummySuperPa"}
	resp2:=map[string]string{"error":database.ErrInvalidCredentials.Error(),"status":controller.Fail}

	//empty data
	third:=&models.Credentials{}
	resp3:=map[string]string{"error":controller.ErrJsonInput}

	//invalid roles
	fourth:=&models.Credentials{"invalidRole","email","pass"}
	resp4:=map[string]string{}

	//data for invalid credentials
	fifth:=&models.Credentials{dummySuperAdmin.Role,"email","dummySuperPa"}

	tests:=[]struct{
		Name string
		data *models.Credentials
		wantStatus int
		wantResp map[string]string
	}{
		{"When an admin is successfully logged in",first,http.StatusOK,resp1},
		{"SuperAdmin with invalid credentials",second,http.StatusUnauthorized,resp2},
		{"with empty fields",third,http.StatusBadRequest,resp3},
		{"For invalid role",fourth,http.StatusNotFound,resp4},
		{"Invalid email and pass",fifth,http.StatusUnauthorized,resp2},
	}
	for _,test :=range tests{
		t.Run(test.Name,func(t *testing.T){
			request:=NewLogInRequest(test.data)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if resp:=response.Body.String();resp!="" {
				assertResponse(t, response.Body.String(), test.wantResp)
			}
		})
	}



}

func TestLogOut(t *testing.T){
	DB,err:=InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer CleanDB(DB)
	router,err:=GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}

	token:=GetSuperToken(router)

	resp1:=map[string]string{"msg":"Logged Out Successfully","status":controller.Success}

	///tests for logout
	testLogout:=[]struct{
		name string
		token string
		wantStatus int
		wantResp map[string]string
		tableName string
	}{
		{"when request is made with token",token,http.StatusOK,resp1,TokenTable},
		{"when token is not sent","",http.StatusBadRequest,nil,""},
	}
	for _,test:=range testLogout{
		t.Run(test.name,func(t *testing.T){
			request:=NewLogOutRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if response.Code==200 {
				assertResponse(t, response.Body.String(), test.wantResp)
				assertTokenEntry(t, DB, test.token, test.tableName)
			}
		})
	}
}



///
func NewLogInRequest(data *models.Credentials) *http.Request{
	jsonData,err:=json.Marshal(data)
	marshalError(err)
	req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(string(jsonData)))
	req.Header.Add("Content-Type", "application/json")
	return req
}
func NewLogOutRequest(token string) *http.Request{
	req, _ := http.NewRequest(http.MethodGet, "/logout",nil)
	req.Header.Add("token",token)
	return req
}

func assertTokenEntry(t *testing.T,db *mysql.MySqlDB,wantToken ,tableName string){
	if tableName==""{
		return
	}
	gotToken:=""
	result,err:=db.Query(fmt.Sprintf("select token from %s where token=?",tableName),wantToken)
	if err!=nil{
		log.Fatalf("can not perform query:%v",err)
	}
	result.Next()
	result.Scan(&gotToken)
	if gotToken==""{
		t.Error("token  not  stored in database")
	}

}