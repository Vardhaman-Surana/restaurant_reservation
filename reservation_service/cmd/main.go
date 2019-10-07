package main

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/queue"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/server"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/tracing"
	"log"
	"os"
	"os/signal"
	"time"
)

const(
	defaultPort="8100"
	defaultDBURL="root:password@tcp(localhost)/restaurant_reservation?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
	defaultRabbitURL="amqp://guest:guest@localhost:5672/"
	)

func main() {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBURL")
	RabbitURL:=os.Getenv("RABBITMQ_URL")
	if port==""{
		port=defaultPort
	}
	if dbURL==""{
		dbURL=defaultDBURL
	}
	if RabbitURL==""{
		RabbitURL=defaultRabbitURL
	}

	_=queue.InitializeQueue(RabbitURL)
	dbMap, err := mysql.NewMysqlDbMap(dbURL)
	if err != nil {
		log.Fatalf("error initiating the db map:%v", err)
	}
	tracer,closer:=tracing.NewTracer(tracing.ServiceName)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	ctx:=context.Background()
	go func(){
		for{
			time.Sleep(1 *time.Minute)
			dbMap.MarkReservationAsDeleted(ctx)
		}
	}()


	s, err := server.NewServer(dbMap,tracer)
	if err != nil {
		panic(err)
	}
	go queue.ConsumeMessage(tracer,dbMap)
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