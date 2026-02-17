package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedData struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Items   []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
}

type OrderEvent struct {
	Event string           `json:"event"`
	Data  OrderCreatedData `json:"data"`
}

type StockHandler func(productID uuid.UUID, quantity int) error

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(host, port, user, password string) (*Consumer, error) {
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

	// Declare order exchange (to bind to it)
	err = ch.ExchangeDeclare("order.exchange", "topic", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue and bind
	_, err = ch.QueueDeclare("order.created.product", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind("order.created.product", "order.created", "order.exchange", false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("RabbitMQ consumer connected")
	return &Consumer{conn: conn, channel: ch}, nil
}

func (c *Consumer) ConsumeOrderCreated(handler StockHandler) {
	msgs, err := c.channel.Consume("order.created.product", "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume: %v", err)
		return
	}

	go func() {
		for msg := range msgs {
			var event OrderEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to unmarshal order event: %v", err)
				continue
			}

			log.Printf("Received order.created: order_id=%s", event.Data.OrderID)

			for _, item := range event.Data.Items {
				productID, err := uuid.Parse(item.ProductID)
				if err != nil {
					log.Printf("Invalid product_id: %s", item.ProductID)
					continue
				}
				if err := handler(productID, item.Quantity); err != nil {
					log.Printf("Failed to decrement stock for product %s: %v", item.ProductID, err)
				}
			}
		}
	}()

	log.Println("Consuming order.created events...")
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
