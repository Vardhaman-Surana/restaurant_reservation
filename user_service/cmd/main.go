package main

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/user_service/pkg/server"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"os"
)

const ServiceName = "userSvc"

func main() {

	port:=os.Getenv("PORT")
	dbURL:=os.Getenv("DBURL")


	dbMap,err:= mysql.NewMysqlDbMap(dbURL)
	if err!=nil{
		log.Fatalf("error initiating the db map:%v",err)
	}
	tracer,closer:=tracing.InitTracer(ServiceName)
	defer closer.Close()

	// create server
	s, err := server.NewServer(dbMap,tracer)
	if err != nil {
		panic(err)
	}
	_,err=s.Start(port)
	if err!=nil{
		panic(err)
	}
}
