package rabbitmq_queue

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"os"
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

func InitializeQueue() *Queue {
	uploadNumTables.Name = queueName
	once.Do(func() {
		log.Println("*********************************")
		log.Println("Inside Once")
		log.Println("*********************************")
		rabbitURL:=os.Getenv("RABBITMQ_URL")
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

func rConnect(url string) *amqp.Connection{
	log.Println("*********************************")
	log.Println(" Creating Connection")
	log.Println("*********************************")
	log.Printf("the url is %v",url)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("trying to reconnect")
		time.Sleep(5 * time.Second)
		return rConnect(url)
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

func SendMessage(resID int,numTables int)error{
	data:=map[string]int{"resID":resID,"numTables":numTables}
	byteData,err:=json.Marshal(data)
	if err!=nil{
		return err
	}
	uploadNumTables.PublishData(byteData)
	log.Printf("\nMessage Sent\n")
	return nil
}
