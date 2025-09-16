package main

import (
	"testing"

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

// TestProducer_Close tests that Close handles nil producer gracefully
func TestProducer_Close(t *testing.T) {
	producer := &Producer{}

	err := producer.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "producer is nil")
}

// TestSendMessage_EmptyValidation tests validation for empty messages
func TestSendMessage_EmptyValidation(t *testing.T) {
	producer := &Producer{
		topic: "test-topic",
	}

	// Test that both empty key and value returns error
	err := producer.SendMessage("", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "both key and value cannot be empty")
}

// TestProducerStruct tests the Producer struct initialization
func TestProducerStruct(t *testing.T) {
	producer := &Producer{
		topic: "test-topic",
	}

	assert.Equal(t, "test-topic", producer.topic)
	assert.Nil(t, producer.producer)
}
