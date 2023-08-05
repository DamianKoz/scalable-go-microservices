package main

import (
	"fmt"
	"listener/event"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitCon, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitCon.Close()

	fmt.Println("Listening for and consuming RabbitMQ messages.")

	consumer, err := event.NewConsumer(rabbitCon)
	if err != nil {
		panic(err)
	}

	err = consumer.Listen([]string{"log.INFO", "log.WARN", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 3 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not ready yet ...")
			counts++
		} else {
			fmt.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		fmt.Println("Backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
