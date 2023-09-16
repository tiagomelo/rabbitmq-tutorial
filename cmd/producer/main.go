package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tiagomelo/rabbitmq-tutorial/producer"
)

const (
	url                = "amqp://guest:guest@localhost:5672/"
	exchangeName       = "my_exchange"
	exchangeKind       = "direct"
	exchangeBindingKey = "upTheIrons"
	queueName          = "testQueue"
)

func main() {
	ctx := context.Background()
	logger := log.New(os.Stdout, "PRODUCER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	// without exchange
	p, err := producer.New(url, queueName, logger)
	// with exchange
	//p, err := producer.New(url, queueName, &log.Logger{}, producer.WithExchange(exchangeName, exchangeKind, exchangeBindingKey))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer p.Close()
	if err := p.Publish(ctx, []byte("test message")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
