package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(d amqp091.Delivery) error

type ConsumerConfig struct {
	QueueName string
	Handler   MessageHandler
}

type Broker struct {
	conn      *amqp091.Connection
	consumers []ConsumerConfig
}

func NewBroker(amqpURL string) (*Broker, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	b := &Broker{
		conn:      conn,
		consumers: []ConsumerConfig{},
	}

	if err := b.declareTopology(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Broker) RegisterConsumer(queueName string, handler MessageHandler) {
	b.consumers = append(b.consumers, ConsumerConfig{
		QueueName: queueName,
		Handler:   handler,
	})
}

func (b *Broker) Start() error {
	var wg sync.WaitGroup
	for _, consumer := range b.consumers {
		wg.Add(1)
		go func(cfg ConsumerConfig) {
			defer wg.Done()
			if err := b.startConsumer(cfg.QueueName, cfg.Handler); err != nil {
				log.Printf("ERROR: RabbitMQ consumer for queue %s stopped: %v", cfg.QueueName, err)
			}
		}(consumer)
	}
	wg.Wait()
	return nil
}

func (b *Broker) Shutdown() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *Broker) startConsumer(queueName string, handler MessageHandler) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("INFO: consuming messages on queue: %s", queueName)
	for d := range msgs {
		err := handler(d)
		if err != nil {
			log.Printf("ERROR: failed to handle message: %v", err)
			_ = d.Nack(false, false)
		} else {
			_ = d.Ack(false)
		}
	}
	return nil
}

func (b *Broker) Reply(d amqp091.Delivery, payload any) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	responseBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(ctx, "", d.ReplyTo, false, false, amqp091.Publishing{
		ContentType:   "application/json",
		CorrelationId: d.CorrelationId,
		Body:          responseBytes,
	})
}

func (b *Broker) declareTopology() error {
	ch, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("gateway_service.v1.events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare("content_service.v1.cv_requests.dlq", true, false, false, false, nil)
	if err != nil {
		return err
	}

	args := amqp091.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "content_service.v1.cv_requests.dlq",
	}
	_, err = ch.QueueDeclare("content_service.v1.cv_requests", true, false, false, false, args)
	if err != nil {
		return err
	}

	err = ch.QueueBind("content_service.v1.cv_requests", "cv.request.*", "gateway_service.v1.events", false, nil)
	if err != nil {
		return err
	}

	log.Println("INFO: RabbitMQ topology declared successfully")
	return nil
}
