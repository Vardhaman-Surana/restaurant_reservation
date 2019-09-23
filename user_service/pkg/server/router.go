package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
)

type Router struct{
	DB database.Database
}

func NewRouter(db database.Database)(*Router,error){
	router := new(Router)
	router.DB = db
	return router,nil
}
func (r *Router)Create() *gin.Engine {
	uc:=controller.NewUserController(r.DB)
	resc:=controller.NewRestaurantController(r.DB)
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