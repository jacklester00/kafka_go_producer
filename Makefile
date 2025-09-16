.PHONY: help test test-verbose test-coverage test-race benchmark clean build run docker-up docker-down docker-logs

# Default target
help:
	@echo "Available commands:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detection"
	@echo "  benchmark     - Run benchmarks"
	@echo "  clean         - Clean build artifacts"
	@echo "  build         - Build the application"
	@echo "  run           - Run the producer"
	@echo "  docker-up     - Start Kafka cluster"
	@echo "  docker-down   - Stop Kafka cluster"
	@echo "  docker-logs   - Show Kafka logs"

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	go test -race ./...

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out coverage.html
	rm -f kafka_go_producer

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

# Start Kafka cluster with UI
docker-up-ui:
	docker-compose --profile ui up -d
	@echo "Waiting for Kafka to be ready..."
	@until docker-compose ps | grep -q "healthy"; do sleep 2; done
	@echo "Kafka is ready! UI available at http://localhost:8080"

# Stop Kafka cluster
docker-down:
	docker-compose down -v

# Stop Kafka cluster and remove volumes
docker-down-clean:
	docker-compose down -v

# Show Kafka logs
docker-logs:
	docker-compose logs -f kafka

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Install linter (if not already installed)
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks (format, lint, test)
check: fmt lint test

# Development workflow: start Kafka, run tests, stop Kafka
dev-test: docker-up test docker-down

# Quick test without Docker (requires external Kafka)
quick-test:
	go test -short ./...

# Integration tests (requires Kafka running)
integration-test: docker-up
	@echo "Running integration tests..."
	go test -tags=integration ./...
	docker-compose down

# Show project status
status:
	@echo "Project: kafka_go_producer"
	@echo "Go version: $(shell go version)"
	@echo "Docker status:"
	@docker-compose ps
	@echo ""
	@echo "Test coverage:"
	@go test -cover ./... | grep -E "(coverage|PASS|FAIL)"

# Send test messages to Kafka (requires Kafka running)
send-test-messages:
	@echo "Sending test messages..."
	@for i in 1 2 3 4 5; do \
		echo "Test message $$i" | docker exec -i kafka kafka-console-producer --bootstrap-server localhost:9092 --topic example-topic; \
	done
	@echo "Test messages sent!"

# Create test topic
create-topic:
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --create --topic example-topic --partitions 3 --replication-factor 1

# List topics
list-topics:
	docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Consume messages from console (for testing)
consume-messages:
	docker exec -it kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic example-topic --from-beginning
