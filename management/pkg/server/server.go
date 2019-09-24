package server

import (
	"errors"
	"github.com/fluent/fluent-logger-golang/fluent"
	_ "github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gin-gonic/gin"
	"github.com/vds/restaurant_reservation/management/pkg/database"
)


type Server struct{
	DB database.Database
	Logger *fluent.Fluent
}


func NewServer(db database.Database,logger *fluent.Fluent)(*Server,error){
	if db == nil {
		return nil, errors.New("server expects a valid database instance")
	}

	return &Server{DB:db,Logger:logger}, nil
}

func(server *Server)Start(port string)(*gin.Engine,error){
	router,err:=NewRouter(server.DB,server.Logger)
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
