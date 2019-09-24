package queue

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/streadway/amqp"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/database"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)
const Tag = "restaurant.reservation"

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

func InitializeQueue(logger *fluent.Fluent) *Queue {
	uploadNumTables.Name = queueName
	er:=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Initializing Queue")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	once.Do(func() {
		log.Println("*********************************")
		log.Println("Inside Once")
		log.Println("*********************************")
		rabbitURL:=os.Getenv("RABBITMQ_URL")
		conn := rConnect(rabbitURL,logger)
		if conn==nil{
			uploadNumTables.Connection=nil
		}else {
			er:=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Connection Created")})
			if er!=nil{
				log.Printf("error in posting log:%v",er)
			}
			log.Println("*********************************")
			log.Println("Connection Created")
			log.Println("*********************************")
			uploadNumTables.Connection = conn
			ch, err := conn.Channel()
			FailOnError(err, "Failed to open a channel")
			log.Println("*********************************")
			log.Println("Channel Created")
			log.Println("*********************************")
			er=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Channel Created")})
			if er!=nil{
				log.Printf("error in posting log:%v",er)
			}
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
			er=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Queue Initialized")})
			if er!=nil{
				log.Printf("error in posting log:%v",er)
			}
		}
	})
	return &uploadNumTables
}

func rConnect(url string,logger *fluent.Fluent) *amqp.Connection{
	log.Println("*********************************")
	log.Println(" Creating Connection")
	log.Println("*********************************")
	log.Printf("the url is %v",url)
	er:=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Creating Connection to url %v",url)})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("trying to reconnect")
		er:=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("trying to reconnect to url %v",url)})
		if er!=nil{
			log.Printf("error in posting log:%v",er)
		}
		time.Sleep(5 * time.Second)
		return rConnect(url,logger)
	}
	return conn
}
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func ConsumeMessage(dbMap database.Database){
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
	}{}
	FailOnError(err,"Failed to register a consumer")
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			err=json.Unmarshal(d.Body,&ResIdAndTables)
			if err!=nil{
				log.Printf("err is %v",err)
			}else{
				log.Println("*********************************")
				log.Println("Got a Message")
				log.Println("*********************************")
				log.Printf("Msg received is %+v",ResIdAndTables)
				err:=dbMap.CreateTablesForRestaurant(ResIdAndTables.ResID,ResIdAndTables.NumTables)
				if err!=nil{
					log.Printf("error is %v",err)
				}
				log.Printf("\nProcessed the message\n")
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
func GetfuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
