package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	dummySuperAdminID="0c647a43-5cef-443e-8688-aaea764d17d2"
 	dummySuperOwnerID="32758fde-4dd5-4635-9bf7-a4d46d6f0629"//created by superadmin
 	dummyOwnerID="451367e3-9b74-4bb6-9157-ac9a2c34da8d"//created by admin
	ExpiredToken ="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImQyZWExMDcxLTMzMjYtNDI1Mi05ODQzLWQ5YWJiMWQ1NDI5MSIsImV4cCI6MTU2ODgwNjY3Nn0.zUmVKIDzOA6YBx5pXMEeBxGOYi74panyuf_mZCiwhBo"
)
func TestRestaurantController(t *testing.T){
	DB,err:=InitDB()
	if err!=nil{
		log.Fatalf("Can not initialize db: %v",err)
	}
	defer CleanDB(DB)
	router,err:=GetRouter()
	if err!=nil{
		log.Fatalf("Can not initialize router: %v",err)
	}

	// tests for get restaurant
	respGetRes1:=[]byte(`[{"id":3,"name":"dummyRestaurant","lat":0.1,"lng":0.2},{"id":4,"name":"dummyRestaurant1","lat":1,"lng":2}]`)
	respGetRes2:=[]byte(`[{"id":3,"name":"dummyRestaurant","lat":0.1,"lng":0.2}]`)
	token:=GetSuperToken(router)
	tokenAdmin:=GetAdminToken(router)
	testGetRestaurants:=[]struct{
		name string
		token string
		resp []byte
	}{
		{"get restaurants with a valid superadmin token",token,respGetRes1},
		{"get restaurants with a token of admin who has not created any restaurant",tokenAdmin,respGetRes2},
	}
	for _,test:=range testGetRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetRestaurantRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,http.StatusOK)
			assertGetResponse(t,response.Body.Bytes(),test.resp)
		})
	}
	//deleting restaurants
	CreateRestaurants(DB)

	respDeleteRes1:=map[string]string{"msg": "Restaurants deleted Successfully"}
	respDeleteRes2:=map[string]string{"error":controller.ErrJsonInput}
	respDeleteRes3:=map[string]string{"error":"Restaurants Deleted Except entry no. 2, 3"}
	respDeleteRes4:=map[string]string{"error":"Restaurants Deleted Except entry no. 1, 2, 3"}
	testDeleteRestaurants:=[]struct{
		name string
		idArr []int
		token string
		wantStatus int
		resp map[string]string
	}{
		{"tests for valid deletion",[]int{5,6},token,http.StatusOK,respDeleteRes1},
		{"empty array of id",nil,token,http.StatusBadRequest,respDeleteRes2},
		{"try to delete invalid id",[]int{7,10,11},token,http.StatusBadRequest,respDeleteRes3},
		{"try to delete with admin token",[]int{6,10,11},tokenAdmin,http.StatusBadRequest,respDeleteRes4},
	}
	for _,test:=range testDeleteRestaurants{
		t.Run(test.name,func(t *testing.T){
			idToDelete:=test.idArr
			request:=NewDeleteRestaurantRequest(test.token,idToDelete)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
		})
	}

	////get available restaurants to add
	respGetAvailableRes1:=[]byte(`[{"id":4,"name":"dummyRestaurant1","lat":1,"lng":2}]`)


	testGetAvailableRestaurants:=[]struct{
		name string
		token string
		wantStatus int
		resp []byte
	}{
		{"get restaurants by superadmin(available here)",token,http.StatusOK,respGetAvailableRes1},
		{"get restaurants by admin(not available here)",tokenAdmin,http.StatusOK,[]byte("[]")},
	}
	for _,test:=range testGetAvailableRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetRestaurantAvailableRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertGetResponse(t,response.Body.Bytes(),test.resp)
		})
	}


	//add restaurant
	respAddRes1:=map[string]string{"error":controller.ErrJsonInput}
	respAddRes2:=map[string]string{"msg":"Restaurant added"}


	testCreateRestaurant:=[]struct{
		name string
		resName string
		lat float64
		lng float64
		wantStatus int
		resp map[string]string
	}{
		{"Create restaurant with empty field","",1.2,5.0,http.StatusBadRequest,respAddRes1},
		{"Create restaurant with valid entries","res100",1.2,5.0,http.StatusOK,respAddRes2},
	}
	for _,test :=range testCreateRestaurant{
		t.Run(test.name,func(t *testing.T){
			request:=NewCreateRestaurantRequest(token,test.resName,test.lat,test.lng)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				assertDbEntryRestaurant(t,DB,test.resName,test.lat,test.lng)
			}
		})
	}

	// updating restaurants

	respUpdateRes1:=map[string]string{"error":controller.ErrJsonInput}
	respUpdateRes2:=map[string]string{"msg":"Restaurant Updated Successfully"}
	respUpdateRes3:=map[string]string{"error":"restaurant does not exist"}
	respUpdateRes4:=map[string]string{"msg":middleware.TokenExpireMessage}


	tokenOwner:=GetOwnerToken(router)
	testUpdateRestaurant:=[]struct{
		name string
		resID int
		resName string
		lat float64
		lng float64
		token string
		wantStatus int
		resp map[string]string
	}{
		{"update restaurants with empty field",3,"",0.1,0.2,token,http.StatusBadRequest,respUpdateRes1},
		{"update restaurants with valid entries",3,"dummyRestaurantUpdate",0.1,0.2,token,http.StatusOK,respUpdateRes2},
		{"update existing restaurant  by it's owner",3,"dummyRestaurantUpdate",0.1,0.2,tokenOwner,http.StatusUnauthorized,nil},
		{"update non existing restaurant by superadmin",1,"dummyRestaurantUpdate",0.1,0.2,token,http.StatusBadRequest,respUpdateRes3},
		{"update with a expired token",3,"dummyRestaurantUpdate",0.1,0.2,ExpiredToken,http.StatusUnauthorized,respUpdateRes4},
	}
	for _,test:=range testUpdateRestaurant{
		t.Run(test.name,func(t *testing.T){
			request:=NewUpdateRestaurantRequest(test.token,test.resID,test.resName,test.lat,test.lng)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if test.resp!=nil{
				assertResponse(t,response.Body.String(),test.resp)
			}
			if response.Code==200{
				assertDbEntryRestaurant(t,DB,test.resName,test.lat,test.lng)
			}
		})
	}
	DB.Exec(`Update restaurants set name="dummyRestaurant" where id=3`)
	//get restaurant for an owner
	respGetOwnerRes1:=[]byte(`[{"id":3,"name":"dummyRestaurant","lat":0.1,"lng":0.2}]`)
	respGetOwnerRes2:=[]byte(`{"error":"can not update owner created by other admin"}`)

	testOwnerRestaurants:=[]struct{
		name string
		token string
		ownerID string
		wantStatus int
		resp []byte
	}{
		{"getting restaurants of an owner by superadmin",token,dummySuperOwnerID,http.StatusOK,respGetOwnerRes1},
		{"getting restaurants of an owner by admin",tokenAdmin,dummySuperOwnerID,http.StatusUnauthorized,respGetOwnerRes2},
		{"getting restaurants of owner by admin created by him",tokenAdmin,dummyOwnerID,http.StatusOK,[]byte("[]")},
	}
	for _,test:=range testOwnerRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetOwnerRestaurantRequest(test.token,test.ownerID)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertGetResponse(t,response.Body.Bytes(),test.resp)
		})
	}

	//add owner to restaurants
	respAddOwnerRes1:=map[string]string{"error":"can not update owner created by other admin"}
	respAddOwnerRes2:=map[string]string{"msg":"Owner assigned restaurants Successfully"}
	respAddOwnerRes3:=map[string]string{"error":controller.ErrJsonInput}
	respAddOwnerRes4:=map[string]string{"error":"owner does not exist"}
	respAddOwnerRes5:=map[string]string{"error":"Restaurants Deleted Except entry no. 1"}


	testAddOwnerRestaurants:=[]struct{
		name string
		token string
		ownerID string
		resID []int
		wantStatus int
		resp map[string]string
	}{
		{"admin trying to add owner to a restaurant not created by him",tokenAdmin,dummySuperOwnerID,[]int{3},http.StatusUnauthorized,respAddOwnerRes1},
		{"superAdmin trying to add owner to a restaurant",token,dummySuperOwnerID,[]int{4},http.StatusOK,respAddOwnerRes2},
		{"empty restaurants field",token,dummySuperOwnerID,nil,http.StatusBadRequest,respAddOwnerRes3},
		{"superAdmin trying to add a non existing owner to a restaurant",token,"1ajk",[]int{3},http.StatusBadRequest,respAddOwnerRes4},
		{"admin trying to add owner to a restaurant not created by him",tokenAdmin,dummyOwnerID,[]int{4},http.StatusBadRequest,respAddOwnerRes5},
		{"admin trying to add owner to a non existing restaurant",tokenAdmin,dummyOwnerID,[]int{15},http.StatusBadRequest,respAddOwnerRes5},
	}
	for _,test:=range testAddOwnerRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewAddOwnerRestaurantRequest(test.token,test.ownerID,test.resID)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				assertDbOwnerAssignRes(t,DB,test.resID,test.ownerID)
			}
		})
	}
	DB.Exec(`update restaurants set owner_id=null where name="dummyRestaurant1"`)

	///get restaurants near by
	respNearByRes1:=[]byte(`[{"name":"dummyRestaurant1"}]`)
	respNearByRes2:=[]byte(`[]`)
	respNearByRes3:=[]byte(`{"error":"Invalid Json Input"}`)

	testGetNearByRestaurants:=[]struct{
		name string
		lat float32
		lng float32
		wantStatus int
		resp []byte
	}{
		{"get restaurants nearby(available)",1.08,2.01,http.StatusOK,respNearByRes1},
		{"get restaurants nearby(not available)",11.08,12.01,http.StatusOK,respNearByRes2},
		{"empty lat and lng(value 0)",0,0,http.StatusBadRequest,respNearByRes3},
	}
	for _,test:=range testGetNearByRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetNearByRestaurants(test.lat,test.lng)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertGetResponse(t,response.Body.Bytes(),test.resp)
		})
	}

}

///
func NewGetRestaurantRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/restaurants",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewCreateRestaurantRequest(token string,name string,lat float64,lng float64) *http.Request{
	restaurant:=models.Restaurant{
		Name: name,
		Lat: lat,
		Lng: lng,
	}
	data,err:=json.Marshal(restaurant)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/manage/restaurants", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewUpdateRestaurantRequest(token string,resID int,name string,lat float64,lng float64) *http.Request{
	restaurant:=models.Restaurant{
		Name: name,
		Lat: lat,
		Lng: lng,
	}
	data,err:=json.Marshal(restaurant)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/restaurants/%d",resID),strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func CreateRestaurants(db *mysql.MySqlDB){
	stmt,_:=db.Prepare("insert into restaurants(name,lat,lng,creator_id,owner_id) values(?,?,?,?,null)")
	stmt.Exec("res10",10,10,dummySuperAdminID)
	stmt.Exec("res20",10,10,dummySuperAdminID)
	stmt.Exec("res30",10,10,dummySuperAdminID)
}
func NewDeleteRestaurantRequest(token string,idArr []int) *http.Request{
	var resID struct {
		IDArr []int	`json:"idArr"`
	}
	resID.IDArr=idArr
	data,err:=json.Marshal(resID)
	if err!=nil{
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/restaurants", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewGetOwnerRestaurantRequest(token string,ownerID string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/manage/owners/%s/restaurants",ownerID),nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewAddOwnerRestaurantRequest(token string,ownerID string,idArr []int) *http.Request{
	var resID struct {
		IDArr []int	`json:"idArr"`
	}
	resID.IDArr=idArr
	data,err:=json.Marshal(resID)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/manage/owners/%s/restaurants",ownerID),strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewGetRestaurantAvailableRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/available/restaurants",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func NewGetNearByRestaurants(lat float32,lng float32) *http.Request{
	location:=map[string]float32{
		"lat":lat,
		"lng":lng,
	}
	data,err:=json.Marshal(location)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, "/restaurantsNearBy",strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}
func assertDbEntryRestaurant(t *testing.T,DB *mysql.MySqlDB,resName string,wantlat,wantlng float64){
	t.Helper()
	var gotLat float64
	var gotLng float64
	rows,err:=DB.Query("select lat,lng from restaurants where name=?",resName)
	if err!=nil{
		fmt.Printf("error performing query to get restauratn:%v",err)
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&gotLat,&gotLng)
	if err!=nil{
		log.Fatalf("error in assigning values to the variables")
	}
	if gotLng!=wantlng && gotLat!=wantlat{
		t.Fatalf("error adding restaurant")
	}
}
func assertDbOwnerAssignRes(t *testing.T,DB *mysql.MySqlDB,resID []int,wantOwner string){
	var gotOwner string
	for _,id:=range resID {
		rows, err := DB.Query("select owner_id from restaurants where id=?", id)
		if err!=nil{
			fmt.Printf("error performing query to get restauratn:%v",err)
		}
		rows.Next()
		err=rows.Scan(&gotOwner)
		if err!=nil{
			log.Fatalf("error in assigning values to the variables")
		}
		if gotOwner!=wantOwner{
			t.Fatalf("Can not assign owner to restaurant")
		}
	}
}
