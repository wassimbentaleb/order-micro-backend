package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type InventoryUpdatedData struct {
	ProductID         string `json:"product_id"`
	QuantityRemaining int    `json:"quantity_remaining"`
	IsLowStock        bool   `json:"is_low_stock"`
}

type InventoryEvent struct {
	Event string               `json:"event"`
	Data  InventoryUpdatedData `json:"data"`
}

type InventoryHandler func(data InventoryUpdatedData)

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

	err = ch.ExchangeDeclare("product.exchange", "topic", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	_, err = ch.QueueDeclare("inventory.updated.order", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind("inventory.updated.order", "inventory.updated", "product.exchange", false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("RabbitMQ consumer connected")
	return &Consumer{conn: conn, channel: ch}, nil
}

func (c *Consumer) ConsumeInventoryUpdated(handler InventoryHandler) {
	msgs, err := c.channel.Consume("inventory.updated.order", "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume: %v", err)
		return
	}

	go func() {
		for msg := range msgs {
			var event InventoryEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to unmarshal inventory event: %v", err)
				continue
			}

			log.Printf("Received inventory.updated: product_id=%s, remaining=%d, low_stock=%v",
				event.Data.ProductID, event.Data.QuantityRemaining, event.Data.IsLowStock)

			handler(event.Data)
		}
	}()

	log.Println("Consuming inventory.updated events...")
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
