package testHelpers

import (
	"bytes"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/models"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/server"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	TestDbURL=`root:password@tcp(localhost)/restaurant_test?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true`
	UserID ="userIdConst"
	ExpiredToken ="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImQyZWExMDcxLTMzMjYtNDI1Mi05ODQzLWQ5YWJiMWQ1NDI5MSIsImV4cCI6MTU2ODgwNjY3Nn0.zUmVKIDzOA6YBx5pXMEeBxGOYi74panyuf_mZCiwhBo"
)
type Claims struct{
	Id string
	jwt.StandardClaims
}

type ReservationAddBody struct {
	ResID     int   `json:"resID"`
	StartTime int64 `json:"startTime"`
}

var (
	GlobalDB *mysql.MysqlDbMap
	OnceDB sync.Once
	OnceRouter sync.Once
	Router *mux.Router
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

func GetRouter()(*mux.Router,error){
	var err error
	OnceRouter.Do(func(){
		initRouter,er:=server.NewRouter(GlobalDB,nil)
		if er!=nil{
			err=er
			Router=nil
		}else {
			Router =initRouter.Create()
			err=er
		}
	})
	return Router,err
}

func CleanDB() {
	_,_=GlobalDB.DbMap.Exec("delete from restaurant_tables")
	_,_=GlobalDB.DbMap.Exec("alter table restaurant_tables AUTO_INCREMENT=1")
	_,_=GlobalDB.DbMap.Exec("delete from reservations")
	_,_=GlobalDB.DbMap.Exec("alter table reservations AUTO_INCREMENT=1")

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

func GetUserToken() string{
	claims:=Claims{Id:UserID}
	token,err:=createToken(&claims)
	if err!=nil{
		log.Fatalf("err in creating user token: %v",err)
	}
	return token
}
func AssertGetResponse(t *testing.T,got []byte,want []byte){
	t.Helper()
	if !bytes.Equal(got,want){
		t.Fatalf("got %s want %s",got,want)
	}
}
func createToken(claims *Claims) (string,error){
	jwtKey:=[]byte("SecretKey")
	expirationTime:=time.Now().Add(60*time.Minute).Unix()
	claims.ExpiresAt=expirationTime
	//remember to change it later
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err!=nil{
		return "",err
	}
	return tokenString,nil
}
func NewGetAvailableReservation(url string,token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, url,nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func NewAddReservationRequest(data *ReservationAddBody,token string)*http.Request{
	body,err:=json.Marshal(data)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/addReservation",strings.NewReader(string(body)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func AssertDbEntryReservation(t *testing.T,data *ReservationAddBody){
	t.Helper()
	var output models.Reservation
	err:=GlobalDB.DbMap.SelectOne(&output,"select * from reservations order by id limit 1")
	if err!=nil{
		panic(err)
	}
	if output.ResID!=data.ResID && output.UserID!=UserID && output.StartTime!=data.StartTime{
		t.Fatalf("error in table entry of the reservation")
	}
}