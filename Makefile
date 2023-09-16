# ==============================================================================
# Help

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==============================================================================
# RabbitMQ

.PHONY: start-rabbitmq
## start-rabbitmq: starts RabbitMQ via docker-compose
start-rabbitmq:
	@ docker-compose up

.PHONY: stop-rabbitmq
## stop-rabbitmq: stops RabbitMQ via docker-compose
stop-rabbitmq:
	@ docker-compose stop

# ==============================================================================
# RabbitMQ producer

.PHONY: producer
## producer: starts RabbitMQ producer
producer:
	@ go run cmd/producer/main.go

# ==============================================================================
# RabbitMQ consumer

.PHONY: consumer
## consumer: starts RabbitMQ consumer
consumer:
	@ go run consumer/main.go

# ==============================================================================
# Tests

.PHONY: test
## test: run unit tests
test:
	@ go test -cover -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report in html format
coverage:
	@ go test -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out