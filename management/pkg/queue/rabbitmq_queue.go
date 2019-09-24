package rabbitmq_queue

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/streadway/amqp"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type Queue struct{
	Name string
	Ch *amqp.Channel
	Connection *amqp.Connection
}
const (
	Tag = "restaurant.management"
	queueName ="UploadNumTables"
)
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
func(q *Queue)PublishData(data []byte){
	err := q.Ch.Publish(
		"",     // exchange
		q.Name, 		// routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
			DeliveryMode:amqp.Persistent,
		})
	log.Println(" Sent ")
	FailOnError(err, "Failed to publish a message")
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func Close(){
	if uploadNumTables.Name!=""{
		uploadNumTables.Connection.Close()
		uploadNumTables.Ch.Close()
		log.Println("queue closed")
	}
}

func SendMessage(resID int,numTables int,logger *fluent.Fluent)error{
	data:=map[string]int{"resID":resID,"numTables":numTables}
	byteData,err:=json.Marshal(data)
	if err!=nil{
		return err
	}
	uploadNumTables.PublishData(byteData)
	log.Printf("\nMessage Sent\n")
	er:=logger.Post(Tag,map[string]string{"infunc":GetfuncName(),"atTime":fmt.Sprintf("%v",time.Now().UnixNano()/1e6),"req":fmt.Sprintf(""),"info":fmt.Sprintf("Message Sent")})
	if er!=nil{
		log.Printf("error in posting log:%v",er)
	}
	return nil
}
func GetfuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
