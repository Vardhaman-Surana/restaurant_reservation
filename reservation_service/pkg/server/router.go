package server

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/controller"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/middleware"
	"net/http"
)

type Router struct{
	DB database.Database
	Tracer opentracing.Tracer
}

func NewRouter(db database.Database,tracer opentracing.Tracer)(*Router,error){
	router := new(Router)
	router.DB = db
	router.Tracer=tracer
	return router,nil
}
func (r *Router)Create() *mux.Router {
	rc:=controller.NewReservationController(r.DB)
	muxRouter:=mux.NewRouter()
	muxRouter.Use(middleware.InitTrace(r.Tracer))

	grp:=muxRouter.PathPrefix("/").Subrouter()
	grp.Use(middleware.AuthMiddleware)
	{
		grp.HandleFunc("/checkAvailability", rc.CheckAvailability).Methods(http.MethodGet)
		grp.HandleFunc("/addReservation", rc.AddReservation).Methods(http.MethodPost)
	}

	return muxRouter
}
