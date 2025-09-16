# Kafka Go Producer Example

A simple, clean example of building a Kafka producer in Go using the Sarama library with Docker.

## Prerequisites

- [Go](https://golang.org/dl/) 1.21+ 
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

## Project Structure

```
kafka_go_producer/
├── README.md              # This file
├── docker-compose.yml     # Kafka cluster setup (KRaft mode)
├── example_producer.go    # Main producer implementation
├── example_producer_test.go # Simple unit tests
├── integration_test.go    # Integration tests (requires Kafka)
├── .github/workflows/test.yml # CI/CD pipeline
├── .golangci.yml         # Code linting rules
├── Makefile              # Development commands
├── go.mod                # Go dependencies
└── .gitignore            # Git ignore rules
```

## Quick Start

### 1. Start Kafka

```bash
# Start Kafka cluster
make docker-up

# Or manually with docker-compose
docker-compose up -d
```

### 2. Run the Producer

```bash
# Install dependencies and run
go mod tidy
go run example_producer.go
```

### 3. Verify Messages

```bash
# In another terminal, consume messages to verify they were sent
make consume-messages

# Or manually with docker
docker exec -it kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic example-topic --from-beginning
```

Your producer will send various types of messages and you should see them in the consumer!

## Configuration

### Producer Settings

The producer is configured to:
- Connect to `localhost:9092`
- Send messages to topic `example-topic`
- Use Snappy compression
- Wait for all in-sync replicas to acknowledge
- Retry up to 5 times on failure
- Flush messages every 500ms

### Customization

Modify these values in `example_producer.go`:

```go
brokers := []string{"localhost:9092"}  // Kafka brokers
topic := "example-topic"               // Topic to produce to
```

## Development Commands

```bash
# Docker operations
make docker-up          # Start Kafka cluster
make docker-down        # Stop Kafka cluster
make docker-logs        # View Kafka logs

# Development
make build              # Build the application
make run                # Build and run producer
make test               # Run tests

# Kafka operations
make consume-messages   # View messages in console
```

## Features

### 1. Simple Message Sending
```go
err := producer.SendMessage("my-key", "my-value")
```

### 2. Messages with Headers
```go
headers := map[string]string{
    "source": "my-app",
    "timestamp": time.Now().Format(time.RFC3339),
}
err := producer.SendMessageWithHeaders("key", "value", headers)
```

### 3. Batch Processing
```go
messages := []MessageData{
    {Key: "key1", Value: "value1"},
    {Key: "key2", Value: "value2"},
}
err := producer.SendBatchMessages(messages)
```

### 4. Streaming Producer
```go
messagesChan := make(chan MessageData, 100)
err := producer.StartBatchProducer(ctx, messagesChan, batchSize, interval)
```

## Testing

### Unit Tests
```bash
go test -v                    # Run unit tests
make test                     # Run via Makefile
```

### Integration Tests
```bash
make integration-test         # Starts Kafka, runs tests, cleans up
go test -tags=integration -v  # Run manually (requires Kafka running)
```

### Benchmarks
```bash
make benchmark               # Run performance benchmarks
go test -bench=. -benchmem   # Run manually
```

## Examples

The producer includes several example functions:

- `ExampleAsyncProducer()` - High-throughput async producer
- `ExampleProducerWithRetry()` - Custom retry logic
- `ExampleProducerWithPartitioning()` - Control message partitioning

## Troubleshooting

**Kafka won't start**: Check `docker-compose logs kafka`

**Producer can't connect**: Verify Kafka is healthy with `docker-compose ps`

**Messages not being sent**: Check topic exists with:
```bash
make list-topics
# or
docker exec -it kafka kafka-topics --bootstrap-server localhost:9092 --list
```

**Reset everything**:
```bash
make docker-down        # Removes all data and containers
make docker-up          # Fresh start
```

## Performance Tips

1. **Use async producer** for high throughput scenarios
2. **Batch messages** when possible to reduce network overhead
3. **Enable compression** (Snappy/LZ4) for better network utilization
4. **Tune batch size and linger time** based on your latency requirements
5. **Use appropriate acknowledgment levels** (acks=all for durability, acks=1 for performance)

## Production Considerations

- **Monitoring**: Add metrics collection (Prometheus/Grafana)
- **Logging**: Implement structured logging with correlation IDs
- **Error Handling**: Implement dead letter queues for failed messages
- **Security**: Configure SASL/SSL for production environments
- **Partitioning**: Design partition keys for even distribution
- **Schema Registry**: Use Avro/Protobuf for message schemas

## Resources

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Sarama Go Client](https://github.com/Shopify/sarama)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Kafka Best Practices](https://kafka.apache.org/documentation/#bestpractices)
