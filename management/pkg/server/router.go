package server

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"net/http"
)


type Router struct{
	db database.Database
	Tracer opentracing.Tracer
	http.Handler
}

func NewRouter(db database.Database,tracer opentracing.Tracer)(*Router,error){
	router := new(Router)
	router.db = db
	router.Tracer=tracer
	return router,nil
}
func (r *Router)Create() *mux.Router{
	//new code
	muxRouter := mux.NewRouter()

	muxRouter.Use(middleware.InitTrace(r.Tracer))
	//Controllers
	regController:=controller.NewRegisterController(r.db)
	loginController:=controller.NewLogInController(r.db)
	resController:=controller.NewRestaurantController(r.db)
	menuController:=controller.NewMenuController(r.db)
	adminController:=controller.NewAdminController(r.db)
	helloworldController:=controller.NewHelloWorldController(r.db)
	ownerController:=controller.NewOwnerController(r.db)


	//Routes

	muxRouter.HandleFunc("/register",regController.Register).Methods(http.MethodPost)
	muxRouter.HandleFunc("/login",loginController.LogIn).Methods(http.MethodPost)
	muxRouter.HandleFunc("/logout",loginController.LogOut).Methods(http.MethodGet)
	muxRouter.HandleFunc("/",helloworldController.SayHello).Methods(http.MethodGet)


	manage:=muxRouter.PathPrefix("/manage").Subrouter()
	manage.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware, middleware.AdminAccessOnly)
	{
		manage.HandleFunc("/owners",ownerController.GetOwners).Methods(http.MethodGet)
		manage.HandleFunc("/owners",ownerController.AddOwner).Methods(http.MethodPost)
		manage.HandleFunc("/owners/{ownerID}",ownerController.EditOwner).Methods(http.MethodPut)
		manage.HandleFunc("/owners",ownerController.DeleteOwners).Methods(http.MethodDelete)
		manage.HandleFunc("/owners/{ownerID}/restaurants",resController.GetOwnerRestaurants).Methods(http.MethodGet)
		manage.HandleFunc("/available/restaurants",resController.GetAvailableRestaurants).Methods(http.MethodGet)
		manage.HandleFunc("/owners/{ownerID}/restaurants",resController.AddOwnerForRestaurants).Methods(http.MethodPost)
		manage.HandleFunc("/restaurants",resController.AddRestaurant).Methods(http.MethodPost)
		manage.HandleFunc("/restaurants",resController.DeleteRestaurants).Methods(http.MethodDelete)

	}
	manageRestaurant:=muxRouter.PathPrefix("/manage").Subrouter()
	manageRestaurant.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware)
	{
		manageRestaurant.HandleFunc("/restaurants",resController.GetRestaurants).Methods(http.MethodGet)

	}
	manageMenu:=muxRouter.PathPrefix("/manage").Subrouter()
	manageMenu.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware)
	manageMenu.Use(middleware.ValidateRestaurantAndCreator(r.db))
	{
		manageMenu.HandleFunc("/restaurants/{resID}",resController.EditRestaurant).Methods(http.MethodPut)
		manageMenu.HandleFunc("/restaurants/{resID}/menu",menuController.GetMenu).Methods(http.MethodGet)
		manageMenu.HandleFunc("/restaurants/{resID}/menu",menuController.AddDishes).Methods(http.MethodPost)
		manageMenu.HandleFunc("/restaurants/{resID}/menu/{dishID}",menuController.EditDish).Methods(http.MethodPut)
		manageMenu.HandleFunc("/restaurants/{resID}/menu",menuController.DeleteDishes).Methods(http.MethodDelete)
	}
	superAdminOnly:=muxRouter.PathPrefix("/manage").Subrouter()
	superAdminOnly.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware, middleware.SuperAdminAccessOnly)
	{
		superAdminOnly.HandleFunc("/admins",adminController.GetAdmins).Methods(http.MethodGet)
		superAdminOnly.HandleFunc("/admins/{adminID}",adminController.EditAdmin).Methods(http.MethodPut)

		superAdminOnly.HandleFunc("/admins",adminController.DeleteAdmins).Methods(http.MethodDelete)
	}
	muxRouter.HandleFunc("/restaurantsNearBy",resController.GetNearBy).Methods(http.MethodGet)
	return muxRouter
}

