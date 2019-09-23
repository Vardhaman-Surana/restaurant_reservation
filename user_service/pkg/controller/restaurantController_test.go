package controller_test

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
	"github.com/vds/restaurant_reservation/user_service/pkg/testHelpers"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const ExpiredToken ="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImQyZWExMDcxLTMzMjYtNDI1Mi05ODQzLWQ5YWJiMWQ1NDI5MSIsImV4cCI6MTU2ODgwNjY3Nn0.zUmVKIDzOA6YBx5pXMEeBxGOYi74panyuf_mZCiwhBo"

func TestRestaurantController(t *testing.T) {
	DB, err := testHelpers.InitDB()
	if err != nil {
		log.Fatalf("Can not initialize db: %v", err)
	}
	defer testHelpers.CleanDB(DB)
	router, err := testHelpers.GetRouter()
	if err != nil {
		log.Fatalf("Can not initialize router: %v", err)
	}

	token:=testHelpers.GetUserToken(router)

	respGetRes1:=[]byte(`[{"id":3,"name":"dummyRestaurant","lat":0.1,"lng":0.2},{"id":4,"name":"dummyRestaurant1","lat":1,"lng":2}]`)

	respGetRes2:=map[string]interface{}{"msg":nil,"error":middleware.TokenExpireErr}


	testGetRestaurants:=[]struct{
		name string
		token string
		wantStatus int
		resp interface{}
	}{
		{"get restaurants with a valid token",token,http.StatusOK,respGetRes1},
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


