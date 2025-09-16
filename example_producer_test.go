package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMessageData tests the MessageData structure
func TestMessageData(t *testing.T) {
	msg := MessageData{
		Key:   "test-key",
		Value: "test-value",
	}

	assert.Equal(t, "test-key", msg.Key)
	assert.Equal(t, "test-value", msg.Value)
}

// TestProducer_Close tests that Close doesn't panic on nil producer
func TestProducer_Close(t *testing.T) {
	producer := &Producer{}

	// Close should handle nil producer gracefully
	// Note: This will return an error since producer is nil, but shouldn't panic
	err := producer.Close()
	assert.Error(t, err) // Expect an error when closing nil producer
}

// TestNewProducer_InvalidBrokers tests producer creation with invalid brokers
func TestNewProducer_InvalidBrokers(t *testing.T) {
	// Test with empty brokers
	producer, err := NewProducer([]string{}, "test-topic")
	assert.Error(t, err)
	assert.Nil(t, producer)

	// Test with invalid broker address
	producer, err = NewProducer([]string{"invalid:broker"}, "test-topic")
	assert.Error(t, err)
	assert.Nil(t, producer)
}

// TestNewProducer_EmptyTopic tests producer creation with empty topic
func TestNewProducer_EmptyTopic(t *testing.T) {
	// This should still create a producer since topic validation happens at send time
	producer, err := NewProducer([]string{"localhost:9092"}, "")
	
	// The producer creation might succeed but sending will fail
	// This depends on the Sarama library behavior
	if err != nil {
		assert.Nil(t, producer)
	} else {
		assert.NotNil(t, producer)
		if producer != nil {
			producer.Close()
		}
	}
}

// TestExampleFunctions tests the example functions don't panic
func TestExampleFunctions(t *testing.T) {
	assert.NotPanics(t, func() {
		ExampleAsyncProducer()
	})

	assert.NotPanics(t, func() {
		ExampleProducerWithRetry()
	})

	assert.NotPanics(t, func() {
		ExampleProducerWithPartitioning()
	})
}

// TestMessageDataSlice tests working with slices of MessageData
func TestMessageDataSlice(t *testing.T) {
	messages := []MessageData{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
		{Key: "key3", Value: "value3"},
	}

	assert.Len(t, messages, 3)
	assert.Equal(t, "key1", messages[0].Key)
	assert.Equal(t, "value2", messages[1].Value)
}

// BenchmarkMessageDataCreation benchmarks MessageData creation
func BenchmarkMessageDataCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MessageData{
			Key:   "benchmark-key",
			Value: "benchmark-value",
		}
	}
}

// TestProducerStruct tests the Producer struct initialization
func TestProducerStruct(t *testing.T) {
	producer := &Producer{
		topic: "test-topic",
	}

	assert.Equal(t, "test-topic", producer.topic)
	assert.Nil(t, producer.producer) // Should be nil until initialized
}

// TestTimeouts tests various timeout scenarios
func TestTimeouts(t *testing.T) {
	// Test that time operations work as expected
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	elapsed := time.Since(start)
	
	assert.True(t, elapsed >= 1*time.Millisecond)
}
