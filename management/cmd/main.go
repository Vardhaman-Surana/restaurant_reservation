package main

import (
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	rabbitmq_queue "github.com/vds/restaurant_reservation/management/pkg/queue"
	"github.com/vds/restaurant_reservation/management/pkg/server"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"log"
	"os"
	"os/signal"
)

const(
	defaultPort="8000"
	defaultDBURL="root:password@tcp(localhost)/restaurant?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
)

func main() {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBURL")

	if port==""{
		port=defaultPort
	}
	if dbURL==""{
		dbURL=defaultDBURL
	}
	// create database instance
	// when not using db4free the restaurant
	db, err := mysql.NewMySqlDB(dbURL)
	if err != nil {
		panic(err)
	}
	tracer,closer:=tracing.NewTracer(tracing.ServiceName)
	defer closer.Close()
	_=rabbitmq_queue.InitializeQueue()

	// create server
	s, err := server.NewServer(db,tracer)
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for sig := range c {
			log.Printf("interrupt signal %v, closing connection",sig)
			rabbitmq_queue.Close()
			log.Printf("queue closed")
			os.Exit(0)
		}
	}()
	_,err=s.Start(port)
	if err!=nil{
		panic(err)
	}
}
