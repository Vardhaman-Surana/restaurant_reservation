package server

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/controller"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/middleware"
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
func (r *Router)Create() *gin.Engine {
	uc:=controller.NewUserController(r.DB,r.Tracer)
	resc:=controller.NewRestaurantController(r.DB,r.Tracer)
	ginRouter:=gin.Default()

	ginRouter.POST("/register",uc.Register)
	ginRouter.POST("/login",uc.LogIn)
	ginRouter.GET("/logout",uc.LogOut)


	grp:=ginRouter.Group("/")
	grp.Use(middleware.TokenValidator(r.Tracer,r.DB),middleware.AuthMiddleware(r.Tracer))
	{
		grp.GET("/restaurants", resc.GetRestaurants)
	}

	return ginRouter
}