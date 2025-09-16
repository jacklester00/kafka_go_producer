//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBroker = "localhost:9092"
	testTopic  = "integration-test-topic"
)

// TestProducerIntegration tests basic producer functionality
func TestProducerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	// Test sending a single message
	err = producer.SendMessage("test-key", "test-value")
	assert.NoError(t, err, "Failed to send message")
}

// TestBatchMessages tests sending multiple messages
func TestBatchMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer, err := NewProducer([]string{testBroker}, testTopic)
	require.NoError(t, err, "Failed to create producer")
	defer producer.Close()

	messages := []MessageData{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	err = producer.SendBatchMessages(messages)
	assert.NoError(t, err, "Failed to send batch messages")
}
