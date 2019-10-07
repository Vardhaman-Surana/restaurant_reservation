package server

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"net/http"
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

func(server *Server)Start(port string)(*mux.Router,error){
	router,err:=NewRouter(server.DB,server.Tracer)
	if err!=nil{
		return nil,err
	}
	r := router.Create()
	err=http.ListenAndServe(":"+port,r)
	if err != nil {
		panic(err)
		return nil,err
	}
	return r,nil
}