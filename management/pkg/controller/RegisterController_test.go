package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	_ "testing"
)

const(
	adminTable="admins"
	superAdminTable="super_admins"
)

func TestRegisterController(t *testing.T){
	DB,err:=InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer CleanDB(DB)
	router,err:=GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}


	//// data preparation for tests
	first:=&models.UserReg{"admin","admin1@gmail.com","admin1","pass1"}
	resp1:=map[string]string{"role":"admin","msg":"Registration Successful","status":"Success",}

	second:=&models.UserReg{"superAdmin","superadmin@gmail.com","superadmin1","superpass1"}
	resp2:=map[string]string{"role":"superAdmin","msg":"Registration Successful","status":"Success",}

	third:=&models.UserReg{"admin","admin1@gmail.com","admin1","pass1"}
	fourth:=&models.UserReg{"superAdmin","superadmin@gmail.com","superadmin1","superpass1"}
	resp3:=map[string]string{"error":database.ErrDupEmail.Error(),"status":"Fail"}

	fifth:=&models.UserReg{"","admin1@gmail.com","admin1","pass1"}
	resp4:=map[string]string{"error":controller.ErrJsonInput,"status":"Fail"}

	tests:= []struct{
		name string
		data *models.UserReg
		wantStatus int
		wantResponse map[string]string
		tableName string
	}{
		{"When an admin is successfully created",first,http.StatusOK,resp1,adminTable},
		{"When a superadmin is successfully created",second,http.StatusOK,resp2,superAdminTable},
		{"duplicate mail for admin",third,http.StatusBadRequest,resp3,""},
		{"duplicate mail for super admin",fourth,http.StatusBadRequest,resp3,""},
		{"Empty Require Field",fifth,http.StatusBadRequest,resp4,""},
	}
	for _,test :=range tests{
		t.Run(test.name,func(t *testing.T){
			request:=NewRegisterRequest(test.data)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if resp:=response.Body.String();resp!="" {
				assertResponse(t, response.Body.String(), test.wantResponse)
			}
			if test.tableName!="" {
				assertDbEntryAdmins(t, DB, test.tableName, test.data)
			}
		})
	}


	data:=&models.UserReg{"AnyOtherRole","name","mail","pass"}
	marshalError(err)
	t.Run("For invalid role",func(t *testing.T){
		request:=NewRegisterRequest(data)
		response:=httptest.NewRecorder()
		router.ServeHTTP(response,request)
		assertStatus(t,response.Code,http.StatusNotFound)
	})
}

//Helpers
func NewRegisterRequest(data *models.UserReg) *http.Request{
	jsonData,err:=json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(string(jsonData)))
	if err!=nil{
		log.Fatal("error forming a registration request")
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}

func CleanDB(db *mysql.MySqlDB) {
	_, _ = db.Query("delete from admins where email_id<>?",dummyAdmin.Email)
	_, _ = db.Query("delete from super_admins where email_id<>?",dummySuperAdmin.Email)
	_, _ = db.Query("delete from restaurants where id<>? and id<>?",3,4)
	_,_=db.Query("alter table restaurants AUTO_INCREMENT=5")
	_, _ = db.Query("delete from dishes where id<>1")
	_,_=db.Query("alter table dishes AUTO_INCREMENT=2")
	_, _ = db.Query("delete from owners where email_id<>? and id<>? ",dummyOwner.Email,dummyOwnerID)
	_,_=db.Query("delete from invalid_tokens")
}
func assertStatus(t *testing.T,got int,want int){
	t.Helper()
	if got!=want{
		t.Fatalf("got status %v want status %v",got,want)
	}
}

func marshalError(err error){
	if err!=nil{
		log.Fatalf("can not marshal data into json:%v",err)
	}
}
func assertResponse(t *testing.T,got string,want map[string]string){
	t.Helper()
	wantJson,err:=json.Marshal(want)
	marshalError(err)
	wantString:=string(wantJson)
	if got!=wantString{
		t.Fatalf("got %v want %v",got,string(wantJson))
	}
}
func assertDbEntryAdmins(t *testing.T,DB *mysql.MySqlDB,tableName string,want *models.UserReg){
	t.Helper()
	got:=&models.UserReg{}
	result,err:=DB.Query(fmt.Sprintf("select email_id,name,password from %s where email_id=?",tableName),want.Email)
	if err!=nil{
		log.Fatalf("can not perform query:%v",err)
	}
	result.Next()
	result.Scan(&got.Email,&got.Name,&got.Password)

	if want.Name==got.Name && want.Email==got.Email &&  encryption.ComparePasswords(got.Password,want.Password){
		return
	}
	t.Errorf("Invalid Data entry got +%v want +%v",got,want)
}