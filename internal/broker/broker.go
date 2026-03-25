package broker

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func Connect(ctx context.Context, amqpURL string) (*Client, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("opening channel: %w", err)
	}

	c := &Client{conn: conn, ch: ch}

	if err := c.declareExchanges(); err != nil {
		_ = c.Close()
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

type MessageHandler func(body []byte) error

func (c *Client) Subscribe(exchange, routingKey, queueName string, handler func(body []byte) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	q, err := c.ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declaring queue %q: %w", queueName, err)
	}

	if err := c.ch.QueueBind(q.Name, routingKey, exchange, false, nil); err != nil {
		return fmt.Errorf("binding queue %q to %q: %w", queueName, exchange, err)
	}

	msgs, err := c.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consuming from %q: %w", queueName, err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}()

	return nil
}
