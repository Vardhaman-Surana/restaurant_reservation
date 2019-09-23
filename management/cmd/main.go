package main

import (
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	rabbitmq_queue "github.com/vds/restaurant_reservation/management/pkg/queue"
	"github.com/vds/restaurant_reservation/management/pkg/server"
	"log"
	"os"
	"os/signal"
)

func main() {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBURL")
	// create database instance
	// when not using db4free the restaurant
	db, err := mysql.NewMySqlDB(dbURL)
	if err != nil {
		panic(err)
	}

	_=rabbitmq_queue.InitializeQueue()

	// create server
	s, err := server.NewServer(db)
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
