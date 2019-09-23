package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const dummyAdminID="a9d68eea-7984-4ca9-a2db-f472d4be2527"

func TestOwnerController(t *testing.T) {
	DB,err:=InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer CleanDB(DB)
	router,err:=GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}
	//test get owners
	token:=GetSuperToken(router)
	tokenAdmin:=GetAdminToken(router)
	tokenOwner:=GetOwnerToken(router)

	respGetOwners1:=[]byte(`[{"id":"32758fde-4dd5-4635-9bf7-a4d46d6f0629","email":"dummySuperOwner@gmail.com","name":"dummySuperOwner"},{"id":"451367e3-9b74-4bb6-9157-ac9a2c34da8d","email":"dummyOwnerAdmin@gmail.com","name":"dummyOwnerAdmin"}]`)
	respGetOwners2:=[]byte(`[{"id":"451367e3-9b74-4bb6-9157-ac9a2c34da8d","email":"dummyOwnerAdmin@gmail.com","name":"dummyOwnerAdmin"}]`)
	testGetOwners:=[]struct{
		name string
		token string
		resp []byte
		wantStatus int
	}{
		{"get owners with a super admin token",token,respGetOwners1,http.StatusOK},
		{"get owners with an admin token where no admins are there",tokenAdmin,respGetOwners2,http.StatusOK},
		{"owner trying to get other owners",tokenOwner,nil,http.StatusUnauthorized},
	}
	for _,test:=range testGetOwners{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetOwnerRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if response.Code==200 {
				assertGetResponse(t, response.Body.Bytes(), test.resp)
			}
		})
	}


	///owner update

	respUpdate1:=map[string]string{"error":controller.ErrJsonInput}
	respUpdate2:=map[string]string{"msg":"Owner updated successfully"}
	respUpdate3:=map[string]string{"error":"owner does not exist"}
	respUpdate4:=map[string]string{"error":"can not update owner created by other admin"}
	respUpdate5:=map[string]string{"error":"email already used try a different one"}
	ownerUpdate:=struct{
		Name string
		Email string
	}{"dummySuperOwnerUpdate","dummySuperOwner@gmail.com"}
	testUpdateOwner:=[]struct{
		name string
		token string
		ownerID string
		ownerEmail string
		ownerName string
		wantStatus int
		resp map[string]string
	}{
		{"update owner with empty field",token,"5d1606c1-0d82-48c4-9bea-6db088e4ad","","",http.StatusBadRequest,respUpdate1},
		{"update existing owner successfully",token,"32758fde-4dd5-4635-9bf7-a4d46d6f0629",ownerUpdate.Email,ownerUpdate.Name,http.StatusOK,respUpdate2},
		{"update non existing owner",token,"32758fde-4dd5-4635-9bf7-a4d46d6f",ownerUpdate.Email,ownerUpdate.Name,http.StatusBadRequest,respUpdate3},
		{"update existing owner with invalid creator",tokenAdmin,"32758fde-4dd5-4635-9bf7-a4d46d6f0629",ownerUpdate.Email,ownerUpdate.Name,http.StatusUnauthorized,respUpdate4},
		{"update non existing by admin",tokenAdmin,"32758fde-4dd5-4635-9",ownerUpdate.Email,ownerUpdate.Name,http.StatusUnauthorized,respUpdate3},
		{"update  owner with a duplicate email",token,"32758fde-4dd5-4635-9bf7-a4d46d6f0629","dummyOwnerAdmin@gmail.com",ownerUpdate.Name,http.StatusBadRequest,respUpdate5},

	}
	for _,test:=range testUpdateOwner{
		t.Run(test.name,func(t *testing.T){
			request:=NewUpdateOwnerRequest(test.token,test.ownerID,test.ownerEmail,test.ownerName)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				assertUpdateOwner(t,DB,test.ownerEmail,test.ownerName)
			}
		})
	}
	// reverting changes
	DB.Exec(`update owners set name="dummySuperOwner",email_id="dummySuperOwner@gmail.com" where name="dummySuperOwnerUpdate"`)

	// deleting owners
	CreateOwners(DB) //creating owners for deletion

	respDeleteOwner1:=map[string]string{"msg":"Owner deleted successfully"}
	respDeleteOwner2:=map[string]string{"error":controller.ErrJsonInput}
	respDeleteOwner3:=map[string]string{"error":"Owners Deleted Except entry no. 1, 2"}

	testDeleteOwners:=[]struct{
		name string
		token string
		idArr []string
		wantStatus int
		resp map[string]string
	}{
		{"tests for valid deletion",token,[]string{"id1","id2","id3"},http.StatusOK,respDeleteOwner1},
		{"tests for valid deletion by admin",tokenAdmin,[]string{"id4"},http.StatusOK,respDeleteOwner1},
		{"empty array of id",token,nil,http.StatusBadRequest,respDeleteOwner2},
		{"try to delete invalid id",token,[]string{"idInvalid","id2Invalid"},http.StatusBadRequest,respDeleteOwner3},
		{"try to delete invalid id by admin",tokenAdmin,[]string{"idInvalid","id2Invalid"},http.StatusBadRequest,respDeleteOwner3},
	}

	for _,test:=range testDeleteOwners{
		t.Run(test.name,func(t *testing.T){
			idToDelete:=test.idArr
			request:=NewDeleteOwnerRequest(test.token,idToDelete)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
		})
	}



	// creating owner
	respCreateOwner1:=map[string]string{"error":controller.ErrJsonInput}
	respCreateOwner2:=map[string]string{"msg":"Owners created successfully"}
	respCreateOwner3:=map[string]string{"error":"email already used try a different one"}

	testCreateOwner:=[]struct{
		name string
		email string
		ownerPass string
		ownerName string
		wantStatus int
		resp map[string]string
	}{
		{"Create owner with empty field","email","","name",http.StatusBadRequest,respCreateOwner1},
		{"Create owner with valid entries","email","password","name",http.StatusOK,respCreateOwner2},
		{"Create owner with duplicate email","email","pass","name",http.StatusBadRequest,respCreateOwner3},
	}
	for _,test:=range testCreateOwner{
		t.Run(test.name,func(t *testing.T){
			request:=NewCreateOwnerRequest(token,test.email,test.ownerName,test.ownerPass)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if response.Code==200 {
				assertResponse(t, response.Body.String(), test.resp)
				assertDbEntryOwner(t,DB,test.email,test.ownerName,test.ownerPass)
			}
		})
	}

}

///
func NewGetOwnerRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/owners",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewUpdateOwnerRequest(token string,ownerID string,email string,userName string) *http.Request{
	data:=fmt.Sprintf(`{"email":"%s","name":"%s"}`,email,userName)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/owners/%s",ownerID),strings.NewReader(data))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewCreateOwnerRequest(token string,email string,userName string,pass string) *http.Request{
	user:=models.OwnerReg{
		Email:    email,
		Name:     userName,
		Password: pass,
	}
	data,err:=json.Marshal(user)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/manage/owners", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func CreateOwners(db *mysql.MySqlDB){
	stmt,_:=db.Prepare("insert into owners(id,email_id,name,password,creator_id) values(?,?,?,?,?)")
	stmt.Exec("id1","email1","name1","pass1","creator1")
	stmt.Exec("id2","email2","name2","pass2","creator2")
	stmt.Exec("id3","email3","name3","pass3","creator3")
	stmt.Exec("id4","email4","name4","pass4",dummyAdminID)
}
func NewDeleteOwnerRequest(token string,idArr []string) *http.Request{
	var ownerID struct {
		IDArr []string	`json:"idArr"`
	}
	ownerID.IDArr=idArr
	data,err:=json.Marshal(ownerID)
	if err!=nil{
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/owners", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func assertUpdateOwner(t *testing.T,DB *mysql.MySqlDB,email,gotName string){
	rows,err:=DB.Query("select name from owners where email_id=?",email)
	var wantName string
	if err!=nil{
		log.Fatalf("error in performing the query")
	}
	defer rows.Close()
	rows.Next()
	rows.Scan(&wantName)
	if gotName!=wantName{
		t.Errorf("owner not updated in the database")
	}
}
func assertDbEntryOwner(t *testing.T,DB *mysql.MySqlDB,email,wantName,pass string){
	rows,err:=DB.Query("select name,password from owners where email_id=?",email)
	var gotName string
	var gotPass string
	if err!=nil{
		log.Fatalf("error in performing the query")
	}
	defer rows.Close()
	rows.Next()
	rows.Scan(&gotName,&gotPass)
	if gotName!=wantName{
		t.Errorf("owner not added in the database: invalid name")
	}
	if !encryption.ComparePasswords(gotPass,pass){
		t.Errorf("owner not added in the database: invalid password")
	}
}