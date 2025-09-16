//go:build integration
// +build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBroker = "localhost:9092"
	testTopic  = "integration-test-topic"
)

// TestProducerIntegration tests the producer against a real Kafka instance
func TestProducerIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create producer
	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	// Test sending a single message
	err = producer.SendMessage("integration-key", "integration-value")
	assert.NoError(t, err, "Failed to send message")
}

// TestProducerWithHeadersIntegration tests sending messages with headers
func TestProducerWithHeadersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	headers := map[string]string{
		"test-header": "test-value",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	err = producer.SendMessageWithHeaders("header-key", "header-value", headers)
	assert.NoError(t, err, "Failed to send message with headers")
}

// TestBatchProducerIntegration tests the batch producer functionality
func TestBatchProducerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	messages := []MessageData{
		{Key: "batch-key-1", Value: "batch-value-1"},
		{Key: "batch-key-2", Value: "batch-value-2"},
		{Key: "batch-key-3", Value: "batch-value-3"},
	}

	err = producer.SendBatchMessages(messages)
	assert.NoError(t, err, "Failed to send batch messages")
}

// TestProducerConsumerRoundTrip tests that messages sent by producer can be consumed
func TestProducerConsumerRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create producer
	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	// Create a simple consumer to verify messages
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer([]string{testBroker}, config)
	require.NoError(t, err, "Failed to create consumer")
	defer consumer.Close()

	// Send a test message
	testKey := "roundtrip-key"
	testValue := "roundtrip-value"
	err = producer.SendMessage(testKey, testValue)
	require.NoError(t, err, "Failed to send test message")

	// Consume the message
	partitionConsumer, err := consumer.ConsumePartition(testTopic, 0, sarama.OffsetNewest)
	require.NoError(t, err, "Failed to create partition consumer")
	defer partitionConsumer.Close()

	// Send another message after consumer is ready
	time.Sleep(100 * time.Millisecond)
	err = producer.SendMessage(testKey, testValue)
	require.NoError(t, err, "Failed to send verification message")

	// Wait for message with timeout
	select {
	case msg := <-partitionConsumer.Messages():
		assert.Equal(t, testKey, string(msg.Key))
		assert.Equal(t, testValue, string(msg.Value))
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// TestBatchProducerWithContext tests the batch producer with context cancellation
func TestBatchProducerWithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messagesChan := make(chan MessageData, 10)
	
	// Start batch producer
	go func() {
		err := producer.StartBatchProducer(ctx, messagesChan, 2, 1*time.Second)
		assert.Error(t, err) // Should error due to context cancellation
	}()

	// Send some messages
	for i := 0; i < 3; i++ {
		messagesChan <- MessageData{
			Key:   "batch-context-key",
			Value: "batch-context-value",
		}
		time.Sleep(300 * time.Millisecond)
	}

	close(messagesChan)
	
	// Wait for context to timeout
	<-ctx.Done()
}

// TestProducerReconnection tests producer behavior when Kafka is temporarily unavailable
func TestProducerReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test assumes Kafka might be temporarily unavailable
	// In a real scenario, you might stop/start Kafka during this test
	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	// Try to send messages - should work if Kafka is available
	for i := 0; i < 5; i++ {
		err := producer.SendMessage("reconnect-key", "reconnect-value")
		if err != nil {
			t.Logf("Expected error during reconnection test: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// BenchmarkProducerThroughput benchmarks producer message throughput
func BenchmarkProducerThroughput(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	if err != nil {
		b.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			err := producer.SendMessage("bench-key", "bench-value")
			if err != nil {
				b.Errorf("Failed to send message: %v", err)
			}
			i++
		}
	})
}

// TestProducerErrorHandling tests various error conditions
func TestProducerErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with invalid topic name (contains invalid characters)
	producer, err := NewProducer([]string{testBroker}, "invalid/topic/name")
	if err == nil && producer != nil {
		defer producer.Close()
		// Try to send a message - this should fail
		err = producer.SendMessage("key", "value")
		assert.Error(t, err, "Expected error when sending to invalid topic")
	}
}
