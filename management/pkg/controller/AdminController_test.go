package controller_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminController(t *testing.T){
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
	tokenAdmin:=GetAdminToken(router)


	respGetAdmins:=[]byte(`[{"id":"a9d68eea-7984-4ca9-a2db-f472d4be2527","email":"dummyAdmin@gmail.com","name":"dummyAdmin"}]`)
	t.Run("get admins with a valid token",func(t *testing.T){
		request:=NewGetAdminRequest(token)
		response:=httptest.NewRecorder()
		router.ServeHTTP(response,request)
		assertStatus(t,response.Code,http.StatusOK)
		assertGetResponse(t,response.Body.Bytes(),respGetAdmins)
	})
	t.Run("get admins with a admin token",func(t *testing.T){
		request:=NewGetAdminRequest(tokenAdmin)
		response:=httptest.NewRecorder()
		router.ServeHTTP(response,request)
		assertStatus(t,response.Code,http.StatusUnauthorized)
	})
	// For update admins
	respupdate1:=map[string]string{"error":controller.ErrJsonInput}
	respupdate2:=map[string]string{"error":"Admin does not exist"}
	respupdate3:=map[string]string{"msg":"admin updated successfully"}


	admin:=struct{
		Name string
		Email string
		ID string
	}{"dummyAdminUpdate","dummyAdmin@gmail.comUpdate","a9d68eea-7984-4ca9-a2db-f472d4be2527"}
	testUpdateAdmin:=[]struct{
		name string
		adminID string
		userName string
		email string
		wantStatus int
		wantResp map[string]string
	}{
		{"update admin with empty field","invalidAdminID","","",http.StatusBadRequest,respupdate1},
		{"update non existing admin","invalidAdminID","name","email",http.StatusBadRequest,respupdate2},
		{"update existing admin",admin.ID,admin.Name,admin.Email,http.StatusOK,respupdate3},
	}
	for _,test:=range testUpdateAdmin{
		t.Run(test.name,func(t *testing.T){
			request:=NewUpdateAdminRequest(token,test.adminID,test.userName,test.email)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.wantResp)
			if response.Code==200{
				assertUpdateAdmin(t,DB,test.email,test.userName)
			}
		})
	}
	//resetting  the changes
	DB.Exec(`update admins set name="dummyAdmin",email_id="dummyAdmin@gmail.com" where name="dummyAdminUpdate"`)

	////Tests for admin deletion

	///create admins for deletion
	CreateAdmins(DB)

	respdelete1:=map[string]string{"msg":"Admins deleted successfully"}
	respdelete2:=map[string]string{"error":controller.ErrJsonInput}
	respdelete3:=map[string]string{"error":"Admins Deleted Except entry no. 1, 2"}

	testDeleteAdmins:=[]struct{
		name string
		idArr []string
		wantStatus int
		wantResp map[string]string
	}{
		{"tests for valid deletion",[]string{"id1","id2","id3"},http.StatusOK,respdelete1},
		{"empty array of id",nil,http.StatusBadRequest,respdelete2},
		{"try to delete invalid id",[]string{"idInvalid","id2Invalid"},http.StatusBadRequest,respdelete3},
	}

	for _,test:=range testDeleteAdmins{
		t.Run(test.name,func(t *testing.T){
			idToDelete:=test.idArr
			request:=NewDeleteAdminRequest(token,idToDelete)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.wantResp)
		})
	}

}

///
func NewGetAdminRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/admins",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewUpdateAdminRequest(token string,adminID string,userName string,email string) *http.Request{
	data:=fmt.Sprintf(`{"email":"%s","name":"%s"}`,email,userName)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/admins/%s",adminID),strings.NewReader(data))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func GetSuperToken(router http.Handler) string{
	cred:=&models.Credentials{dummySuperAdmin.Role,dummySuperAdmin.Email,dummySuperAdmin.Password}
	request:=NewLogInRequest(cred)
	response:=httptest.NewRecorder()
	router.ServeHTTP(response,request)
	token:=response.Header().Get("token")
	fmt.Println("**************************")
	fmt.Printf("token is %v",token)
	fmt.Println("**************************")
	return token
}

func GetAdminToken(router http.Handler) string{
	cred:=&models.Credentials{dummyAdmin.Role,dummyAdmin.Email,dummyAdmin.Password}
	request:=NewLogInRequest(cred)
	response:=httptest.NewRecorder()
	router.ServeHTTP(response,request)
	token:=response.Header().Get("token")
	return token
}
func GetOwnerToken(router http.Handler) string{
	cred:=&models.Credentials{"owner",dummyOwner.Email,dummyOwner.Password}
	request:=NewLogInRequest(cred)
	response:=httptest.NewRecorder()
	router.ServeHTTP(response,request)
	token:=response.Header().Get("token")
	return token
}

func CreateAdmins(db *mysql.MySqlDB){
	stmt,_:=db.Prepare("insert into admins(id,email_id,name,password) values(?,?,?,?)")
	stmt.Exec("id1","email1","name1","pass1")
	stmt.Exec("id2","email2","name2","pass2")
	stmt.Exec("id3","email3","name3","pass3")
}
func NewDeleteAdminRequest(token string,idArr []string) *http.Request{
	var adminID struct {
		IDArr []string	`json:"idArr"`
	}
	adminID.IDArr=idArr
	data,err:=json.Marshal(adminID)
	if err!=nil{
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/admins", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func assertGetResponse(t *testing.T,got []byte,want []byte){
	t.Helper()
	if !bytes.Equal(got,want){
		t.Fatalf("got %s want %s",got,want)
	}
}
func assertUpdateAdmin(t *testing.T,DB *mysql.MySqlDB,email ,wantName string){
 	rows,err:=DB.Query("select name from admins where email_id=?",email)
 	var gotName string
 	if err!=nil{
 		log.Fatalf("error in performing the query")
	}
 	defer rows.Close()

 	rows.Next()
 	rows.Scan(&gotName)
 	if gotName!=wantName{
 		t.Errorf("admins not updated in the database")
	}
}