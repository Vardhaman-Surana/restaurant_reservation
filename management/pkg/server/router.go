package server

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/controller"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
)


type Router struct{
	db database.Database
	Logger *fluent.Fluent
}

func NewRouter(db database.Database,logger *fluent.Fluent)(*Router,error){
	router := new(Router)
	router.db = db
	router.Logger=logger
	return router,nil
}
func (r *Router)Create() *gin.Engine {
	ginRouter:=gin.Default()

	//Controllers
	regController:=controller.NewRegisterController(r.db,r.Logger)
	loginController:=controller.NewLogInController(r.db,r.Logger)
	resController:=controller.NewRestaurantController(r.db,r.Logger)
	menuController:=controller.NewMenuController(r.db,r.Logger)
	adminController:=controller.NewAdminController(r.db,r.Logger)
	helloworldController:=controller.NewHelloWorldController(r.db,r.Logger)


	ownerController:=controller.NewOwnerController(r.db,r.Logger)
	//Routes
	ginRouter.POST("/register",regController.Register)
	ginRouter.POST("/login",loginController.LogIn)
	ginRouter.GET("/logout",loginController.LogOut)
	ginRouter.GET("/",helloworldController.SayHello)


	manage:=ginRouter.Group("/manage")
	manage.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware, middleware.AdminAccessOnly)
	{
		manage.GET("/owners",ownerController.GetOwners)
		manage.POST("/owners",ownerController.AddOwner)
		manage.PUT("/owners/:ownerID",ownerController.EditOwner)
		manage.DELETE("/owners",ownerController.DeleteOwners)
		manage.GET("/owners/:ownerID/restaurants",resController.GetOwnerRestaurants)
		manage.GET("/available/restaurants",resController.GetAvailableRestaurants)
		manage.POST("/owners/:ownerID/restaurants",resController.AddOwnerForRestaurants)


		manage.POST("/restaurants",resController.AddRestaurant)
		manage.DELETE("/restaurants",resController.DeleteRestaurants)

	}
	manageRestaurant:=ginRouter.Group("/manage")
	manageRestaurant.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware)
	{
		manageRestaurant.GET("/restaurants",resController.GetRestaurants)

	}
	manageMenu:=ginRouter.Group("/manage")
	manageMenu.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware)
	manageMenu.Use(middleware.ValidateRestaurantAndCreator(r.db))
	{
		manageMenu.PUT("/restaurants/:resID",resController.EditRestaurant)

		manageMenu.GET("/restaurants/:resID/menu",menuController.GetMenu)
		manageMenu.POST("/restaurants/:resID/menu",menuController.AddDishes)
		manageMenu.PUT("/restaurants/:resID/menu/:dishID",menuController.EditDish)
		manageMenu.DELETE("/restaurants/:resID/menu",menuController.DeleteDishes)
	}
	superAdminOnly:=ginRouter.Group("/manage")
	superAdminOnly.Use(middleware.TokenValidator(r.db),middleware.AuthMiddleware, middleware.SuperAdminAccessOnly)
	{
		superAdminOnly.GET("/admins",adminController.GetAdmins)
		superAdminOnly.PUT("/admins/:adminID",adminController.EditAdmin)
		superAdminOnly.DELETE("/admins",adminController.DeleteAdmins)
	}
	ginRouter.GET("/restaurantsNearBy",resController.GetNearBy)

	return ginRouter
}