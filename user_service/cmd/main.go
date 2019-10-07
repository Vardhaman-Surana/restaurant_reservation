package main

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/user_service/pkg/fireBaseAuth"
	"github.com/vds/restaurant_reservation/user_service/pkg/server"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"log"
	"os"
)

const(
	defaultPort="8200"
	defaultDBURL="root:password@tcp(localhost)/restaurant?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
)
func main() {

	port:=os.Getenv("PORT")
	dbURL:=os.Getenv("DBURL")
	if port==""{
		port=defaultPort
	}
	if dbURL==""{
		dbURL=defaultDBURL
	}

	tracer,closer:=tracing.NewTracer(tracing.ServiceName)
	defer closer.Close()


	dbMap,err:= mysql.NewMysqlDbMap(dbURL)
	if err!=nil{
		log.Fatalf("error initiating the db map:%v",err)
	}

	err=fireBaseAuth.InitFireBase()
	if err!=nil{
		log.Fatalf("err in initializing app")
	}

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
