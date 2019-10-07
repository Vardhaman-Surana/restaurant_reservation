package queue

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"github.com/streadway/amqp"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/tracing"
	"log"
	"sync"
	"time"
)

type Queue struct{
	Name string
	Ch *amqp.Channel
	Connection *amqp.Connection
}

const queueName ="UploadNumTables"

var(
	uploadNumTables Queue
	once     sync.Once
)

func InitializeQueue(rabbitURL string) *Queue {
	uploadNumTables.Name = queueName
	once.Do(func() {
		log.Println("*********************************")
		log.Println("Inside Once")
		log.Println("*********************************")
		conn := rConnect(rabbitURL)
		if conn==nil{
			uploadNumTables.Connection=nil
		}else {
			log.Println("*********************************")
			log.Println("Connection Created")
			log.Println("*********************************")
			uploadNumTables.Connection = conn
			ch, err := conn.Channel()
			FailOnError(err, "Failed to open a channel")
			log.Println("*********************************")
			log.Println("Channel Created")
			log.Println("*********************************")
			_, err = ch.QueueDeclare(
				uploadNumTables.Name, // name
				true,
				false,
				false,
				false,
				nil,
			)
			FailOnError(err, "Failed to declare a queue")
			log.Println("rabbitmq connected")
			uploadNumTables.Ch = ch
		}
	})
	return &uploadNumTables
}

func rConnect(url string) *amqp.Connection {
	log.Println("*********************************")
	log.Println(" Creating Connection")
	log.Println("*********************************")
	log.Printf("the url is %v", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("trying to reconnect")
		time.Sleep(5 * time.Second)
		return rConnect(url)
	}
	return conn
}
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func ConsumeMessage(tracer opentracing.Tracer,dbMap database.Database){
	msgs, err := uploadNumTables.Ch.Consume(
		uploadNumTables.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	ResIdAndTables:=struct{
		ResID int `json:"resID"`
		NumTables int `json:"numTables"`
		ReqID string `json:"reqID"`
	}{}
	FailOnError(err,"Failed to register a consumer")
	forever := make(chan bool)
	go func() {
		for d := range msgs {

			err=json.Unmarshal(d.Body,&ResIdAndTables)
			if err!=nil{
				log.Printf("err is %v",err)
			}else{
				span := tracer.StartSpan("process_message")
				tags:=tracing.TraceTags{FuncName:"ConsumeMessage",ServiceName:tracing.ServiceName,RequestID:ResIdAndTables.ReqID}
				span.SetBaggageItem("requestID",ResIdAndTables.ReqID)
				tracing.SetTags(span,tags)
				log.Println("*********************************")
				log.Println("Got a Message")
				log.Println("*********************************")
				newCtx:=context.Background()
				newCtx=opentracing.ContextWithSpan(newCtx,span)


				_,err:=dbMap.CreateTablesForRestaurant(newCtx,ResIdAndTables.ResID,ResIdAndTables.NumTables)
				if err!=nil{
					log.Printf("error is %v",err)
				}
				log.Printf("\nProcessed the message\n")
				span.Finish()
			}
		}
	}()
	<-forever
}

func Close(){
	if uploadNumTables.Name!=""{
		uploadNumTables.Connection.Close()
		uploadNumTables.Ch.Close()
		log.Println("queue closed")
	}
}

