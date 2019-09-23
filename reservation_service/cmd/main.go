package main

import (
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/queue"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/server"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBURL")
	RabbitURL:=os.Getenv("RABBITMQ_URL")
	_=queue.InitializeQueue(RabbitURL)
	dbMap, err := mysql.NewMysqlDbMap(dbURL)
	if err != nil {
		log.Fatalf("error initiating the db map:%v", err)
	}

	go func(){
		for{
			time.Sleep(1 *time.Minute)
			dbMap.MarkReservationAsDeleted()
		}
	}()


	s, err := server.NewServer(dbMap)
	if err != nil {
		panic(err)
	}
	go queue.ConsumeMessage(dbMap)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for sig := range c {
			log.Printf("interrupt signal %v, closing connection",sig)
			queue.Close()
			log.Printf("queue closed")
			os.Exit(0)
		}
	}()

	_,err=s.Start(port)
	if err!=nil{
		panic(err)
	}
}