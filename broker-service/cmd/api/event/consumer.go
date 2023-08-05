package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data "`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,       // Queue Name
			s,            // topic
			"logs_topic", //exchange
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for m := range messages {
			var payload Payload
			err := json.Unmarshal(m.Body, &payload)
			if err != nil {
				fmt.Println(err)
			}
			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message on [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			fmt.Println(err)
		}

	case "auth":
		// Here could be the authentication

	default:
		err := logEvent(payload)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func logEvent(entry Payload) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest(http.MethodPost, logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
