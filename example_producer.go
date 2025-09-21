package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

// ProducerError represents a custom error type for producer operations
type ProducerError struct {
	Operation string
	Err       error
}

// Error implements error interface
func (e *ProducerError) Error() string {
	return fmt.Sprintf("producer %s failed: %v", e.Operation, e.Err)
}

// Config holds configuration for the Kafka producer
type Config struct {
	Brokers       []string
	Topic         string
	BatchSize     int
	FlushInterval time.Duration
}

// Producer represents a Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer creates a new Kafka producer with default configuration
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := Config{
		Brokers:       brokers,
		Topic:         topic,
		BatchSize:     3,
		FlushInterval: 500 * time.Millisecond,
	}
	return NewProducerWithConfig(config)
}

// NewProducerWithConfig creates a new Kafka producer with custom configuration
func NewProducerWithConfig(config Config) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_1_0                 // Use appropriate Kafka version
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack
	saramaConfig.Producer.Retry.Max = 5                    // Retry up to 5 times to produce the message
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Compression = sarama.CompressionSnappy // Compress messages
	saramaConfig.Producer.Flush.Frequency = config.FlushInterval

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, &ProducerError{
			Operation: "create",
			Err:       err,
		}
	}

	return &Producer{
		producer: producer,
		topic:    config.Topic,
	}, nil
}

// SendMessage sends a single message to Kafka
func (p *Producer) SendMessage(key, value string) error {
	if key == "" && value == "" {
		return &ProducerError{
			Operation: "send",
			Err:       fmt.Errorf("both key and value cannot be empty"),
		}
	}

	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return &ProducerError{
			Operation: "send",
			Err:       err,
		}
	}

	log.Printf("Message sent successfully - Topic: %s, Partition: %d, Offset: %d", p.topic, partition, offset)
	return nil
}

// SendMessageWithHeaders sends a message with custom headers to Kafka
func (p *Producer) SendMessageWithHeaders(key, value string, headers map[string]string) error {
	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	// Add headers if provided
	if headers != nil {
		message.Headers = make([]sarama.RecordHeader, 0, len(headers))
		for k, v := range headers {
			message.Headers = append(message.Headers, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return &ProducerError{
			Operation: "send_with_headers",
			Err:       err,
		}
	}

	log.Printf("Message with headers sent successfully - Topic: %s, Partition: %d, Offset: %d", p.topic, partition, offset)
	return nil
}

// SendBatchMessages sends multiple messages in batch to Kafka
func (p *Producer) SendBatchMessages(messages []MessageData) error {
	for _, msg := range messages {
		err := p.SendMessage(msg.Key, msg.Value)
		if err != nil {
			return &ProducerError{
				Operation: "send_batch",
				Err:       fmt.Errorf("failed to send batch message (key: %s): %w", msg.Key, err),
			}
		}
	}
	return nil
}

// MessageData represents a message to be sent to Kafka
type MessageData struct {
	Key   string
	Value string
}

// StartBatchProducer starts a producer that sends messages at regular intervals.
// It batches messages up to batchSize and sends them either when the batch is full
// or when the interval timer expires.
func (p *Producer) StartBatchProducer(ctx context.Context, messages chan MessageData, batchSize int, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	batch := make([]MessageData, 0, batchSize)

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				// Channel is closed, send any remaining messages and exit
				if len(batch) > 0 {
					if err := p.SendBatchMessages(batch); err != nil {
						log.Printf("Error sending final batch: %v", err)
					}
				}
				log.Println("Message channel closed, batch producer stopping")
				return nil
			}

			batch = append(batch, msg)

			// Send batch when it reaches the desired size
			if len(batch) >= batchSize {
				if err := p.SendBatchMessages(batch); err != nil {
					log.Printf("Error sending batch: %v", err)
				}
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			// Send any remaining messages in batch on timer
			if len(batch) > 0 {
				if err := p.SendBatchMessages(batch); err != nil {
					log.Printf("Error sending timed batch: %v", err)
				}
				batch = batch[:0] // Reset batch
			}

		case <-ctx.Done():
			// Send any remaining messages before shutting down
			if len(batch) > 0 {
				if err := p.SendBatchMessages(batch); err != nil {
					log.Printf("Error sending final batch: %v", err)
				}
			}
			return ctx.Err()
		}
	}
}

// Close closes the producer and releases resources
func (p *Producer) Close() error {
	if p.producer == nil {
		return &ProducerError{
			Operation: "close",
			Err:       fmt.Errorf("producer is nil"),
		}
	}

	if err := p.producer.Close(); err != nil {
		return &ProducerError{
			Operation: "close",
			Err:       err,
		}
	}
	return nil
}

func main() {
	// Configuration
	brokers := []string{"localhost:9092"} // Change to your Kafka broker addresses
	topic := "example-topic"              // Change to your topic name

	// Create producer
	producer, err := NewProducer(brokers, topic)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Example 1: Send individual messages
	log.Println("Sending individual messages...")
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("Hello Kafka! Message %d", i)

		if err := producer.SendMessage(key, value); err != nil {
			log.Printf("Failed to send message %d: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Example 2: Send message with headers
	log.Println("Sending message with headers...")
	headers := map[string]string{
		"source":    "example-producer",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0",
	}
	if err := producer.SendMessageWithHeaders("header-key", "Message with headers", headers); err != nil {
		log.Printf("Failed to send message with headers: %v", err)
	}

	// Example 3: Batch producer
	log.Println("Starting batch producer...")
	messagesChan := make(chan MessageData, 100)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := producer.StartBatchProducer(ctx, messagesChan, 3, 2*time.Second); err != nil {
			log.Printf("Batch producer error: %v", err)
		}
	}()

	// Send some messages to the batch producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(messagesChan)

		for i := 0; i < 10; i++ {
			select {
			case messagesChan <- MessageData{
				Key:   fmt.Sprintf("batch-key-%d", i),
				Value: fmt.Sprintf("Batch message %d", i),
			}:
				log.Printf("Queued batch message %d", i)
			case <-ctx.Done():
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Wait for either shutdown signal or all work to complete
	go func() {
		wg.Wait()
		log.Println("All producer tasks completed, shutting down...")
		cancel()
	}()

	select {
	case <-signals:
		log.Println("Received shutdown signal, stopping producer...")
		cancel()
	case <-ctx.Done():
		log.Println("Producer tasks completed")
	}

	// Wait for any remaining goroutines to finish
	wg.Wait()
	log.Println("Producer stopped gracefully")
}

// Example usage functions for different scenarios

// ExampleAsyncProducer shows how to use async producer for higher throughput
func ExampleAsyncProducer() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.Producer.RequiredAcks = sarama.WaitForLocal // Only wait for the local commit to succeed
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Printf("Failed to create async producer: %v", err)
		return
	}
	defer producer.Close()

	// Handle successes and errors
	go func() {
		for success := range producer.Successes() {
			log.Printf("Message sent successfully: partition=%d offset=%d", success.Partition, success.Offset)
		}
	}()

	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to send message: %v", err)
		}
	}()

	// Send messages
	for i := 0; i < 10; i++ {
		message := &sarama.ProducerMessage{
			Topic: "example-topic",
			Key:   sarama.StringEncoder(fmt.Sprintf("async-key-%d", i)),
			Value: sarama.StringEncoder(fmt.Sprintf("Async message %d", i)),
		}
		producer.Input() <- message
	}

	log.Println("Example async producer - messages sent")
}

// ExampleProducerWithRetry shows how to implement retry logic
func ExampleProducerWithRetry() {
	// This would be implemented with exponential backoff
	// and custom retry logic for failed messages
	log.Println("Example producer with retry - implement retry logic with backoff")
}

// ExampleProducerWithPartitioning shows how to control message partitioning
func ExampleProducerWithPartitioning() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner // Use hash partitioner

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Printf("Failed to create producer: %v", err)
		return
	}
	defer producer.Close()

	// Send message to specific partition
	message := &sarama.ProducerMessage{
		Topic:     "example-topic",
		Partition: 0, // Send to partition 0
		Key:       sarama.StringEncoder("partition-key"),
		Value:     sarama.StringEncoder("Message for specific partition"),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Printf("Failed to send partitioned message: %v", err)
		return
	}

	log.Printf("Partitioned message sent - Partition: %d, Offset: %d", partition, offset)
}
