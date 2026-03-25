package broker_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/broker"
)

func testRabbitMQURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	return url
}

func TestConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()
}

func TestConnect_InvalidURL(t *testing.T) {
	_, err := broker.Connect(context.Background(), "amqp://bad:bad@localhost:1/")
	if err == nil {
		t.Fatal("expected error for bad RabbitMQ URL, got nil")
	}
}

func TestPublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	received := make(chan []byte, 1)

	err = c.Subscribe(broker.ExchangeEvents, "events.proj-1.*", "test-events-queue", func(body []byte) error {
		received <- body
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	msg := []byte(`{"type":"task.completed","task_id":"t-1"}`)
	if err := c.Publish(ctx, broker.ExchangeEvents, "events.proj-1.task_completed", msg); err != nil {
		t.Fatalf("Publish() error: %v", err)
	}

	select {
	case got := <-received:
		if string(got) != string(msg) {
			t.Errorf("got %q, want %q", string(got), string(msg))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestTopicRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	received := make(chan []byte, 2)

	err = c.Subscribe(broker.ExchangeLogs, "logs.proj-1.*", "test-logs-wildcard", func(body []byte) error {
		received <- body
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	// Publish to two different agent IDs under the same project
	_ = c.Publish(ctx, broker.ExchangeLogs, "logs.proj-1.agent-a", []byte(`{"log":"from agent-a"}`))
	_ = c.Publish(ctx, broker.ExchangeLogs, "logs.proj-1.agent-b", []byte(`{"log":"from agent-b"}`))

	for i := 0; i < 2; i++ {
		select {
		case <-received:
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for message %d", i+1)
		}
	}
}

func TestSubscribe_HandlerError_NacksMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	callCount := 0
	handlerErr := fmt.Errorf("handler error")

	// Subscribe with a handler that returns an error.
	err = c.Subscribe(broker.ExchangeEvents, "events.nack-test.*", "test-nack-queue", func(body []byte) error {
		callCount++
		return handlerErr
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	msg := []byte(`{"type":"test"}`)
	if err := c.Publish(ctx, broker.ExchangeEvents, "events.nack-test.agent", msg); err != nil {
		t.Fatalf("Publish() error: %v", err)
	}

	// Give the consumer time to receive and nack the message.
	time.Sleep(500 * time.Millisecond)

	// The handler should have been called at least once.
	if callCount == 0 {
		t.Error("expected handler to be called, got 0 calls")
	}
}

func TestClose_DoubleClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	// First close should succeed.
	if err := c.Close(); err != nil {
		t.Fatalf("first Close() error: %v", err)
	}

	// Second close on an already-closed connection should not panic.
	// It may return an error, which is acceptable.
	_ = c.Close()
}

func TestPublish_ToAllExchanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := broker.Connect(ctx, testRabbitMQURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	exchanges := []string{
		broker.ExchangeEvents,
		broker.ExchangeLogs,
		broker.ExchangeCommands,
	}

	for _, exchange := range exchanges {
		t.Run(exchange, func(t *testing.T) {
			err := c.Publish(ctx, exchange, "test.routing.key", []byte(`{"type":"test"}`))
			if err != nil {
				t.Errorf("Publish() to exchange %q error: %v", exchange, err)
			}
		})
	}
}
