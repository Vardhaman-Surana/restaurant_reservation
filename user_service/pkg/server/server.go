package server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
)

type Server struct{
	DB database.Database
	Tracer opentracing.Tracer
}


func NewServer(db database.Database,tracer opentracing.Tracer)(*Server,error){
	if db == nil {
		return nil, errors.New("server expects a valid database instance")
	}
	return &Server{DB:db,Tracer:tracer}, nil
}

func(server *Server)Start(port string)(*gin.Engine,error){
	router,err:=NewRouter(server.DB,server.Tracer)
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
