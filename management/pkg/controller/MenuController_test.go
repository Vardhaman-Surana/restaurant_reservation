package controller_test

import (
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

func TestMenuController(t *testing.T){
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

	// get menu test
	respGetMenu1:=[]byte(`[{"id":1,"name":"dish1","price":100}]`)
	respGetMenu2:=map[string]string{"error":"restaurant does not exist"}
	respGetMenu3:=[]byte("[]")
	testGetMenu:=[]struct{
		name string
		token string
		resID int
		wantStatus int
		resp interface{}
	}{
		{"for an existing restaurant",token,3,http.StatusOK,respGetMenu1},
		{"for a non existing restaurant",token,10,http.StatusBadRequest,respGetMenu2},
		{"when no items in menu",token,4,http.StatusOK,respGetMenu3},
	}
	for _,test:=range testGetMenu{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetMenuRequest(token,test.resID)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			if response.Code==200{
				assertGetResponse(t,response.Body.Bytes(),test.resp.([]byte))
			}else{
				assertResponse(t,response.Body.String(),test.resp.(map[string]string))
			}
		})
	}


	//tests for deleting dishes
	respDishDelete1:=map[string]string{"msg":"Dishes deleted successfully"}
	respDishDelete2:=map[string]string{"error":controller.ErrJsonInput}
	respDishDelete3:=map[string]string{"error":"Dishes Deleted Except entry no. 1, 2"}

	CreateDishes(DB)
	testDeleteDishes:=[]struct{
		name string
		resID int
		idArr []int
		wantStatus int
		resp map[string]string
	}{
		{"tests for valid deletion",3,[]int{2,3,4},http.StatusOK,respDishDelete1},
		{"empty array of id",3,nil,http.StatusBadRequest,respDishDelete2},
		{"try to delete invalid id",3,[]int{10,11},http.StatusBadRequest,respDishDelete3},
	}
	for _,test:=range testDeleteDishes{
		t.Run(test.name,func(t *testing.T){
			idToDelete:=test.idArr
			request:=NewDeleteDishRequest(token,test.resID,idToDelete)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
		})
	}



	//tokenAdmin:=GetAdminToken(router)
	///test to add dishes to a restaurant
	respAddDish1:=map[string]string{"msg":"Dishes Added to menu successfully"}
	respAddDish2:=map[string]string{"error":"restaurant does not exist"}

	testAddDishes:=[]struct{
		name string
		resID int
		dishes []models.Dish
		wantStatus int
		resp map[string]string
	}{
		{"Add dishes successfully",3,[]models.Dish{{"dish1",100.0}},http.StatusOK,respAddDish1},
		{"Adding dishes for a non existing restaurant",10,[]models.Dish{{"dish1",100.0},{"dish2",200.0}},http.StatusBadRequest,respAddDish2},
	}
	for _,test:=range testAddDishes{
		t.Run(test.name,func(t *testing.T){
			request:=NewAddDishesRequest(token,test.resID,test.dishes)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				assertDbEntryDishes(t,DB,test.dishes[0])
			}
		})
	}


	//tests for updating a dish
	respUpdateDish1:=map[string]string{"error":controller.ErrJsonInput}
	respUpdateDish2:=map[string]string{"msg":"Dish Updated successfully"}
	respUpdateDish3:=map[string]string{"error":"dish does not exist"}

	testUpdateDish:=[]struct{
		name string
		token string
		resID int
		dishID int
		dishName string
		dishPrice float32
		wantStatus int
		resp map[string]string
	}{
		{"with empty fields",token,3,1,"",10.0,http.StatusBadRequest,respUpdateDish1},
		{"update an existing dish",token,3,1,"dish1",100.0,http.StatusOK,respUpdateDish2},
		{"when dish id does not exist",token,3,10,"dish1Update",20.0,http.StatusBadRequest,respUpdateDish3},
	}
	for _,test:=range testUpdateDish{
		t.Run(test.name,func(t *testing.T){
			request:=NewUpdateDishRequest(token,test.resID,test.dishID,test.dishName,test.dishPrice)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
			assertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				assertDbUpdateDish(t,DB,test.dishID,test.dishName,test.dishPrice)
			}
		})
	}
	DB.Exec(`Update dishes set name="dish1",price=100.00 where id=1`)
}


//////
func NewAddDishesRequest(token string,resID int,dishes []models.Dish) *http.Request{
	data,err:=json.Marshal(dishes)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost,fmt.Sprintf("/manage/restaurants/%d/menu",resID), strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewGetMenuRequest(token string,resID int) *http.Request{
	req, _ := http.NewRequest(http.MethodGet,fmt.Sprintf("/manage/restaurants/%d/menu",resID), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewUpdateDishRequest(token string,resID,dishID int,dishName string,dishPrice float32) *http.Request{
	dish:=models.Dish{
		Name:  dishName,
		Price: dishPrice,
	}
	data,err:=json.Marshal(dish)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/restaurants/%d/menu/%d",resID,dishID),strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func CreateDishes(db *mysql.MySqlDB){
	stmt,_:=db.Prepare("insert into dishes(name,price,res_id) values(?,?,?)")
	stmt.Exec("dish10",100,3)
	stmt.Exec("dish20",200,3)
	stmt.Exec("dish30",300,3)
}
func NewDeleteDishRequest(token string,resID int,idArr []int) *http.Request{
	var dishID struct {
		IDArr []int	`json:"idArr"`
	}
	dishID.IDArr=idArr
	data,err:=json.Marshal(dishID)
	if err!=nil{
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/manage/restaurants/%d/menu",resID), strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func assertDbEntryDishes(t *testing.T,DB *mysql.MySqlDB,dish models.Dish){
	rows,err:=DB.Query("select price from dishes where name=?",dish.Name)
	var gotPrice float32
	if err!=nil{
		log.Fatalf("error in performing the query")
	}
	defer rows.Close()
	rows.Next()
	rows.Scan(&gotPrice)
	if gotPrice!=dish.Price{
		t.Errorf("dish not added in the database")
	}
}
func assertDbUpdateDish(t *testing.T,DB *mysql.MySqlDB,id int,wantName string,wantPrice float32){
	rows,err:=DB.Query("select name,price from dishes where id=?",id)
	var gotName string
	var gotPrice float32
	if err!=nil{
		log.Fatalf("error in performing the query:%v",err)
	}
	defer rows.Close()
	rows.Next()
	rows.Scan(&gotName,&gotPrice)
	if gotName!=wantName && gotPrice!=wantPrice{
		t.Fatalf("dish not updated successfully")
	}
}