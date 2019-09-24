package main

import (
	"github.com/fluent/fluent-logger-golang/fluent"
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	/*
	Testing log output in a file
	 */
	/*f, err := os.OpenFile("/Users/vds/reservationService.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	*/

	// using fluent logger
	logger, err := fluent.New(fluent.Config{FluentPort: 24224, FluentHost: "127.0.0.1"})
	if err != nil {
		log.Println(err)
	}
	defer logger.Close()

	_=queue.InitializeQueue(logger)
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


	s, err := server.NewServer(dbMap,logger)
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