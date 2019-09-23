package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/controller"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/middleware"
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
	rc:=controller.NewReservationController(r.DB)
	ginRouter:=gin.Default()

	grp:=ginRouter.Group("/")
	grp.Use(middleware.AuthMiddleware)
	{
		grp.Use(middleware.AuthMiddleware).GET("/checkAvailability", rc.CheckAvailability)
		grp.Use(middleware.AuthMiddleware).POST("/addReservation", rc.AddReservation)
	}

	return ginRouter
}
