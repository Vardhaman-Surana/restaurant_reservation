package server

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
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
	muxRouter:=mux.NewRouter()

	uc:=controller.NewUserController(r.DB)
	resc:=controller.NewRestaurantController(r.DB)
	muxRouter.Use(middleware.InitTrace(r.Tracer))

	muxRouter.HandleFunc("/register",uc.Register).Methods(http.MethodPost)
	muxRouter.HandleFunc("/login",uc.LogIn).Methods(http.MethodPost)


	grp:=muxRouter.PathPrefix("/").Subrouter()
	grp.Use(middleware.TokenValidator(r.DB),middleware.AuthMiddleware)
	{
		grp.HandleFunc("/restaurants", resc.GetRestaurants).Methods(http.MethodGet)
		grp.HandleFunc("/logout",uc.LogOut).Methods(http.MethodGet)

	}

	return muxRouter
}