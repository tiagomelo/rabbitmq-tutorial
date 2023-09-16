package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	url                = "amqp://guest:guest@localhost:5672/"
	exchangeName       = "my_exchange"
	exchangeKind       = "direct"
	exchangeBindingKey = "upTheIrons"
	queueName          = "testQueue"
)

func main() {
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		panic(err)
	}

	if err := ch.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		true,  // durable
		false, // autoDelete
		false, // Internal
		false, // noWait
		nil,   // args,
	); err != nil {
		panic(err)
	}
	if err := ch.QueueBind(
		queueName,
		exchangeBindingKey,
		exchangeName,
		false, // noWait
		nil,   // args
	); err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
