package main

import (
	"github.com/fluent/fluent-logger-golang/fluent"
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
	/*f, err := os.OpenFile("/Users/vds/userService.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	 */

	//using fluent logger
	logger, err := fluent.New(fluent.Config{FluentPort: 24224, FluentHost: "127.0.0.1"})
	if err != nil {
		log.Println(err)
	}
	defer logger.Close()

	dbMap,err:= mysql.NewMysqlDbMap(dbURL)
	if err!=nil{
		log.Fatalf("error initiating the db map:%v",err)
	}
	// create server
	s, err := server.NewServer(dbMap,logger)
	if err != nil {
		panic(err)
	}
	_,err=s.Start(port)
	if err!=nil{
		panic(err)
	}
}
