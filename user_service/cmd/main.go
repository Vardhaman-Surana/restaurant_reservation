package main

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/user_service/pkg/server"
	"log"
	"os"
)

func main() {

	port:=os.Getenv("PORT")
	dbURL:=os.Getenv("DBURL")

	dbMap,err:= mysql.NewMysqlDbMap(dbURL)
	if err!=nil{
		log.Fatalf("error initiating the db map:%v",err)
	}
	// create server
	s, err := server.NewServer(dbMap)
	if err != nil {
		panic(err)
	}
	_,err=s.Start(port)
	if err!=nil{
		panic(err)
	}
}
