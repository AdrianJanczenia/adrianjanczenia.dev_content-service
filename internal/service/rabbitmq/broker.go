package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/registry"
	"github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(d amqp091.Delivery) (replyPayload any, err error)

type ConsumerConfig struct {
	QueueName     string
	Handler       MessageHandler
	ConsumerCount int
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
	return &Broker{
		conn:      conn,
		consumers: []ConsumerConfig{},
	}, nil
}

func (b *Broker) DeclareTopology(cfg registry.RabbitMQTopologyConfig) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	for _, ex := range cfg.Exchanges {
		if err := ch.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, false, false, false, nil); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
	}

	for _, q := range cfg.Queues {
		args := amqp091.Table{}
		if q.DLQ != "" {
			args["x-dead-letter-exchange"] = ""
			args["x-dead-letter-routing-key"] = q.DLQ
		}
		if _, err := ch.QueueDeclare(q.Name, q.Durable, false, false, false, args); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.Name, err)
		}
	}

	for _, bind := range cfg.Bindings {
		queue, ok := cfg.Queues[bind.QueueKey]
		if !ok {
			return fmt.Errorf("queue with key '%s' not found in config for binding", bind.QueueKey)
		}
		if err := ch.QueueBind(queue.Name, bind.RoutingKey, bind.Exchange, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queue.Name, bind.Exchange, err)
		}
	}

	log.Println("INFO: RabbitMQ topology declared successfully")
	return nil
}

func (b *Broker) RegisterConsumer(queueName string, count int, handler MessageHandler) {
	b.consumers = append(b.consumers, ConsumerConfig{
		QueueName:     queueName,
		Handler:       handler,
		ConsumerCount: count,
	})
}

func (b *Broker) Start() error {
	var wg sync.WaitGroup
	for _, consumer := range b.consumers {
		for i := 0; i < consumer.ConsumerCount; i++ {
			wg.Add(1)
			go func(cfg ConsumerConfig) {
				defer wg.Done()
				if err := b.startConsumer(cfg.QueueName, cfg.Handler); err != nil {
					log.Printf("ERROR: RabbitMQ consumer for queue %s stopped: %v", cfg.QueueName, err)
				}
			}(consumer)
		}
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
		replyPayload, err := handler(d)
		if err != nil {
			_ = d.Nack(false, false)
			continue
		}

		if d.ReplyTo != "" {
			err = b.reply(d, replyPayload)
			if err != nil {
				log.Printf("ERROR: failed to send reply: %v", err)
				_ = d.Nack(false, false)
				continue
			}
		}
		_ = d.Ack(false)
	}

	return nil
}

func (b *Broker) reply(d amqp091.Delivery, payload any) error {
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
