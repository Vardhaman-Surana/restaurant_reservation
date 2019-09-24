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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	/*
	Directing log output to a file
	 */
	f, err := os.OpenFile("/Users/vds/userService.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)


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
