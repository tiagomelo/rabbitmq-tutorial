# rabbitmq-tutorial

This sample project aims to show how to write an idiomatic, testable [RabbitMQ](www.rabbitmq.com) producer.

## producer

The producer can be used to either publish messages to a given exchange or not.

Read more about message routing [here](https://www.rabbitmq.com/tutorials/tutorial-four-go.html).

### without exchange

```
logger := log.New(os.Stdout, "PRODUCER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

p, err := producer.New(url, queueName, logger)
if err != nil {
    // ...
}
defer p.Close()

if err := p.Publish(ctx, []byte("test message")); err != nil {
    // ...
}
```

### with exchange

```
logger := log.New(os.Stdout, "PRODUCER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

p, err := producer.New(url, queueName, logger)
if err != nil {
    // ...
}
defer p.Close()

if err := p.Publish(ctx, []byte("test message")); err != nil {
    // ...
}
```

## running it

### RabbitMQ

```
make start-rabbitmq
```

### consumer

This is a sample consumer. It is not optmized and not recommended for production env; it is here just to show the producer working.

```
make consumer
```

### producer

It will run the producer without an exchange configuration.

```
make producer
```

## RabbitMQ console

`http://localhost:15672/`

user and pass are `guest`.

## running tests

```
make test
```

To see a coverage report:

```
make coverage
```