package rabbitmq

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	rdb     *redis.Client
}

type GenericEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

type EventHandlers struct {
	OnUserRegistered    func(userID, username, email string)
	OnOrderCreated      func(orderID, userID string)
	OnOrderCompleted    func(orderID, userID string)
	OnProductOutOfStock func(productID, productName string)
}

func NewConsumer(host, port, user, password string, rdb *redis.Client) (*Consumer, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare all exchanges we need to bind to
	for _, exchange := range []string{"user.exchange", "order.exchange", "product.exchange"} {
		err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	// Declare and bind queues
	queues := []struct {
		name       string
		routingKey string
		exchange   string
	}{
		{"user.registered.notify", "user.registered", "user.exchange"},
		{"order.created.notify", "order.created", "order.exchange"},
		{"order.completed.notify", "order.completed", "order.exchange"},
		{"product.outofstock.notify", "product.out_of_stock", "product.exchange"},
	}

	for _, q := range queues {
		_, err = ch.QueueDeclare(q.name, true, false, false, false, nil)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		err = ch.QueueBind(q.name, q.routingKey, q.exchange, false, nil)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}
	}

	log.Println("RabbitMQ consumer connected")
	return &Consumer{conn: conn, channel: ch, rdb: rdb}, nil
}

func (c *Consumer) isDuplicate(event GenericEvent, body []byte) bool {
	if c.rdb == nil {
		return false
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(body))
	key := fmt.Sprintf("dedup:%s:%s", event.Event, hash)

	set, err := c.rdb.SetNX(context.Background(), key, "1", 24*time.Hour).Result()
	if err != nil {
		log.Printf("Dedup check failed: %v", err)
		return false
	}

	// SetNX returns true if the key was set (i.e., first time seeing this event)
	return !set
}

func (c *Consumer) StartConsuming(handlers EventHandlers) {
	c.consumeQueue("user.registered.notify", func(event GenericEvent, body []byte) {
		handlers.OnUserRegistered(
			getString(event.Data, "user_id"),
			getString(event.Data, "username"),
			getString(event.Data, "email"),
		)
	})

	c.consumeQueue("order.created.notify", func(event GenericEvent, body []byte) {
		handlers.OnOrderCreated(
			getString(event.Data, "order_id"),
			getString(event.Data, "user_id"),
		)
	})

	c.consumeQueue("order.completed.notify", func(event GenericEvent, body []byte) {
		handlers.OnOrderCompleted(
			getString(event.Data, "order_id"),
			getString(event.Data, "user_id"),
		)
	})

	c.consumeQueue("product.outofstock.notify", func(event GenericEvent, body []byte) {
		handlers.OnProductOutOfStock(
			getString(event.Data, "product_id"),
			getString(event.Data, "product_name"),
		)
	})

	log.Println("All notification consumers started")
}

func (c *Consumer) consumeQueue(queueName string, handler func(GenericEvent, []byte)) {
	msgs, err := c.channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume from %s: %v", queueName, err)
		return
	}

	go func() {
		for msg := range msgs {
			var event GenericEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to unmarshal event from %s: %v", queueName, err)
				continue
			}

			if c.isDuplicate(event, msg.Body) {
				log.Printf("Skipping duplicate event from %s: %s", queueName, event.Event)
				continue
			}

			log.Printf("Received event from %s: %s", queueName, event.Event)
			handler(event, msg.Body)
		}
	}()

	log.Printf("Consuming from %s...", queueName)
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
