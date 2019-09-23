package server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
)

type Server struct{
	DB database.Database
}


func NewServer(data database.Database)(*Server,error){
	if data == nil {
		return nil, errors.New("server expects a valid database instance")
	}
	return &Server{DB:data}, nil
}

func(server *Server)Start(port string)(*gin.Engine,error){
	router,err:=NewRouter(server.DB)
	if err!=nil{
		return nil,err
	}
	r := router.Create()
	err=r.Run(":"+port)
	if err != nil {
		panic(err)
		return nil,err
	}
	return r,nil
}