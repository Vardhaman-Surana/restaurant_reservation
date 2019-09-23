package controller_test

import (
	"fmt"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/controller"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/testHelpers"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckAvailability(t *testing.T){
	DB, err := testHelpers.InitDB()
	if err != nil {
		log.Fatalf("Can not initialize db: %v", err)
	}
	defer testHelpers.CleanDB()
	router, err := testHelpers.GetRouter()
	if err != nil {
		log.Fatalf("Can not initialize router: %v", err)
	}
	UserToken:=testHelpers.GetUserToken()

	// add data before test
	DB.CreateTablesForRestaurant(10,2)
	timestring:=fmt.Sprintf("%v",time.Now().Add(1 * time.Minute).Unix())


	reqUrl1:="/checkAvailability?resID=10&startTime="+timestring
	resp1:=map[string]interface{}{"error":"Token expired please login again","msg":nil}

	reqUrl2:="/checkAvailability"
	resp2:=map[string]interface{}{"msg":nil,"error":controller.ErrQueryParamNotFound+" resID startTime"}

	reqUrl3:=`/checkAvailability?resID="50"&startTime="1568905200"`
	resp3:=map[string]interface{}{"msg":nil,"error":controller.ErrInvalidParamType+" startTime resID"}

	reqUrl4:="/checkAvailability?resID=50&startTime=1568905200"
	resp4:=map[string]interface{}{"msg":controller.ReservationNotAvailableMessage,"error":nil}

	resp5:=map[string]interface{}{"msg":controller.ReservationAvailableMessage+"2","error":nil}

	reqUrl6:=`/checkAvailability?resID=50&startTime=12345`
	resp6:=map[string]interface{}{"msg":nil,"error":"entered startTime is of the past"}
	testAvailability:=[]struct{
		name string
		token string
		ReqUrl string
		wantStatus int
		resp map[string]interface{}
	}{
		{"request with an expired token",testHelpers.ExpiredToken,reqUrl1,http.StatusUnauthorized,resp1},
		{"request without query parameters",UserToken,reqUrl2,http.StatusBadRequest,resp2},
		{"request with wrong type query parameters",UserToken,reqUrl3,http.StatusBadRequest,resp3},
		{"request for a restaurant where reservation not available",UserToken,reqUrl4,http.StatusOK,resp4},
		{"request for a restaurant where tables available",UserToken,reqUrl1,http.StatusOK,resp5},
		{"request for time before now",UserToken,reqUrl6,http.StatusNotAcceptable,resp6},

	}
	for _,test:=range testAvailability{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewGetAvailableReservation(test.ReqUrl,test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			if response!=nil{
				testHelpers.AssertResponse(t,response.Body.String(),test.resp)
			}
		})
	}

}

func TestAddReservation(t *testing.T){
	DB, err := testHelpers.InitDB()
	if err != nil {
		log.Fatalf("Can not initialize db: %v", err)
	}
	defer testHelpers.CleanDB()
	router, err := testHelpers.GetRouter()
	if err != nil {
		log.Fatalf("Can not initialize router: %v", err)
	}
	// creating data for db
	DB.CreateTablesForRestaurant(10,2)

	UserToken:=testHelpers.GetUserToken()

	data1:=testHelpers.ReservationAddBody{ResID:0, StartTime:0}
	resp1:=map[string]interface{}{"msg":nil,"error":controller.ErrEmptyFields+" resID startTime"}

	data2:=testHelpers.ReservationAddBody{ResID:50, StartTime:1234}
	resp2:=map[string]interface{}{"msg":nil,"error":"entered startTime is of the past"}

	data3:=testHelpers.ReservationAddBody{ResID:50, StartTime:time.Now().Add(1 * time.Minute).Unix()}
	resp3:=map[string]interface{}{"msg":nil,"error":controller.ReservationNotAvailableMessage}

	data4:=testHelpers.ReservationAddBody{ResID:10, StartTime:time.Now().Add(1 * time.Minute).Unix()}
	resp4:=map[string]interface{}{"msg":controller.ReservationSuccessMessage,"error":nil,"resvID":1}

	testAddReservation:=[]struct{
		name string
		token string
		Data *testHelpers.ReservationAddBody
		wantStatus int
		resp map[string]interface{}
	}{
		{"with zero values",UserToken,&data1,http.StatusBadRequest,resp1},
		{"with pastTime",UserToken,&data2,http.StatusNotAcceptable,resp2},
		{"when reservation not available for the restaurant",UserToken,&data3,http.StatusNotAcceptable,resp3},
		{"successful reservation",UserToken,&data4,http.StatusOK,resp4},
	}
	for _,test:=range testAddReservation{
		t.Run(test.name,func(t *testing.T){
			request:=testHelpers.NewAddReservationRequest(test.Data,test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			testHelpers.AssertStatus(t,response.Code,test.wantStatus)
			testHelpers.AssertResponse(t,response.Body.String(),test.resp)
			if response.Code==200{
				testHelpers.AssertDbEntryReservation(t,test.Data)
			}
		})
	}

}


