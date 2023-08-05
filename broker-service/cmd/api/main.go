package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitCon, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitCon.Close()

	app := Config{
		Rabbit: rabbitCon,
	}

	fmt.Printf("Starting broker service on Port %s\n", webPort)

	//define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
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