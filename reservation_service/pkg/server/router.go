package server

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/controller"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/middleware"
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
	rc:=controller.NewReservationController(r.DB,r.Logger)
	ginRouter:=gin.Default()

	grp:=ginRouter.Group("/")
	grp.Use(middleware.AuthMiddleware)
	{
		grp.Use(middleware.AuthMiddleware).GET("/checkAvailability", rc.CheckAvailability)
		grp.Use(middleware.AuthMiddleware).POST("/addReservation", rc.AddReservation)
	}

	return ginRouter
}
