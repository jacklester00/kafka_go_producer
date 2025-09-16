.PHONY: help test build run docker-up docker-down docker-logs consume-messages

# Default target
help:
	@echo "Available commands:"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the application"
	@echo "  run           - Run the producer"
	@echo "  docker-up     - Start Kafka cluster"
	@echo "  docker-down   - Stop Kafka cluster"
	@echo "  docker-logs   - Show Kafka logs"
	@echo "  consume-messages - View messages in console"

# Run tests
test:
	go test ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the application
build:
	go build -o kafka_go_producer .

# Run the producer
run: build
	./kafka_go_producer

# Start Kafka cluster
docker-up:
	docker-compose up -d
	@echo "Waiting for Kafka to be ready..."
	@for i in $$(seq 1 30); do \
		if docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then \
			echo "Kafka is ready!"; \
			break; \
		fi; \
		echo "Attempt $$i: Kafka not ready yet, waiting..."; \
		sleep 2; \
	done

# Stop Kafka cluster
docker-down:
	docker-compose down -v

# Show Kafka logs
docker-logs:
	docker-compose logs -f kafka

# Consume messages from console
consume-messages:
	docker exec -it kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic example-topic --from-beginning
