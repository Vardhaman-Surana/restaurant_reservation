package server

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
)

type Router struct{
	DB database.Database
	Logger *fluent.Fluent
}

func NewRouter(db database.Database,logger *fluent.Fluent)(*Router,error){
	router := new(Router)
	router.DB = db
	router.Logger=logger
	return router,nil
}
func (r *Router)Create() *gin.Engine {
	uc:=controller.NewUserController(r.DB,r.Logger)
	resc:=controller.NewRestaurantController(r.DB,r.Logger)
	ginRouter:=gin.Default()

	ginRouter.POST("/register",uc.Register)
	ginRouter.POST("/login",uc.LogIn)
	ginRouter.GET("/logout",uc.LogOut)


	grp:=ginRouter.Group("/")
	grp.Use(middleware.TokenValidator(r.DB),middleware.AuthMiddleware)
	{
		ginRouter.Use(middleware.AuthMiddleware).GET("/restaurants", resc.GetRestaurants)
	}

	return ginRouter
}