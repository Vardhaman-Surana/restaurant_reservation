package testHelpers

import (
"bytes"
"encoding/json"
"github.com/gin-gonic/gin"
"github.com/vds/restaurant_reservation/user_service/pkg/database/mysql"
"github.com/vds/restaurant_reservation/user_service/pkg/models"
"github.com/vds/restaurant_reservation/user_service/pkg/server"
"log"
"net/http"
"net/http/httptest"
	"strings"
	"sync"
"testing"
)

const (
	TestDbURL=`root:password@tcp(localhost)/restaurant_test?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true`
	TokenTable="invalid_tokens"
	)

var (
	GlobalDB *mysql.MysqlDbMap
	OnceDB sync.Once
	OnceRouter sync.Once
	Router *gin.Engine
)
func InitDB()(*mysql.MysqlDbMap,error){
	var err error
	OnceDB.Do(func(){
		db,er:=mysql.NewMysqlDbMap(TestDbURL)
		GlobalDB = db
		err=er
	})
	return GlobalDB,err
}

func GetRouter()(*gin.Engine,error){
	var err error
	OnceRouter.Do(func(){
		initRouter,er:=server.NewRouter(GlobalDB)
		if er!=nil{
			err=er
			Router=nil
		}else {
			Router = initRouter.Create()
			err=er
		}
	})
	return Router,err
}

func CleanDB(db *mysql.MysqlDbMap) {
	_,_=db.DbMap.Exec("delete from users where email<>?","vardhaman@gmail.com")
	_,_=db.DbMap.Exec("delete from invalid_tokens")
}

func AssertStatus(t *testing.T,got int,want int){
	t.Helper()
	if got!=want{
		t.Fatalf("got status %v want status %v",got,want)
	}
}

func AssertResponse(t *testing.T,got string,want map[string]interface{}){
	t.Helper()
	wantJson,err:=json.Marshal(want)
	marshalError(err)
	wantString:=string(wantJson)
	if got!=wantString{
		t.Fatalf("got %v want %v",got,string(wantJson))
	}
}

func marshalError(err error){
	if err!=nil{
		log.Fatalf("can not marshal data into json:%v",err)
	}
}

func GetUserToken(router http.Handler) string{
	cred:=&models.User{Email:"vardhaman@gmail.com",Password:"password"}
	request:=NewLogInRequest(cred)
	response:=httptest.NewRecorder()
	router.ServeHTTP(response,request)
	token:=response.Header().Get("token")
	return token
}
func AssertGetResponse(t *testing.T,got []byte,want []byte){
	t.Helper()
	if !bytes.Equal(got,want){
		t.Fatalf("got %s want %s",got,want)
	}
}

func NewRegisterRequest(data *models.User) *http.Request{
	jsonData,err:=json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(string(jsonData)))
	if err!=nil{
		log.Fatal("error forming a registration request")
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}


func NewLogInRequest(data *models.User) *http.Request{
	jsonData,err:=json.Marshal(data)
	marshalError(err)
	req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(string(jsonData)))
	req.Header.Add("Content-Type", "application/json")
	return req
}

func AssertUserEntry(t *testing.T,db *mysql.MysqlDbMap,data *models.User){
	var count int
	err:=db.DbMap.SelectOne(&count,"select count(*) from users where email=?",data.Email)
	if err!=nil{
		panic(err)
	}
	if count!=1{
		t.Fatalf("error in storing the user in db")
	}
}
func NewGetRestaurantRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/restaurants",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewLogOutRequest(token string) *http.Request{
	req, _ := http.NewRequest(http.MethodGet, "/logout",nil)
	req.Header.Add("token",token)
	return req
}