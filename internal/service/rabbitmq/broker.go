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

type replyPublisher interface {
	PublishWithDeferredConfirmWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) (*amqp091.DeferredConfirmation, error)
}

type MessageHandler func(ctx context.Context, d amqp091.Delivery) (replyPayload any, err error)

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
	conn, err := amqp091.DialConfig(amqpURL, amqp091.Config{Dial: amqp091.DefaultDial(3 * time.Second)})
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

func (b *Broker) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)

	for _, consumer := range b.consumers {
		for i := 0; i < consumer.ConsumerCount; i++ {
			wg.Add(1)
			go func(cfg ConsumerConfig) {
				defer wg.Done()
				if err := b.startConsumer(ctx, cfg.QueueName, cfg.Handler); err != nil {
					select {
					case errCh <- fmt.Errorf("RabbitMQ consumer %s: %w", cfg.QueueName, err):
					default:
					}
					cancel()
				}
			}(consumer)
		}
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (b *Broker) Shutdown() error {
	if b.conn != nil {
		return b.conn.Close()
	}

	return nil
}

func (b *Broker) startConsumer(ctx context.Context, queueName string, handler MessageHandler) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return err
	}

	returns := ch.NotifyReturn(make(chan amqp091.Return, 1))
	go func() {
		for returned := range returns {
			log.Printf("WARN: RabbitMQ reply was unroutable (code %d)", returned.ReplyCode)
		}
	}()

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("INFO: consuming messages on queue: %s", queueName)
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("delivery channel closed")
			}

			hCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
			replyPayload, err := handler(hCtx, d)
			if err != nil {
				cancel()
				if nackErr := d.Nack(false, false); nackErr != nil {
					return nackErr
				}
				continue
			}

			if d.ReplyTo != "" {
				err = b.replyConfirmed(hCtx, ch, d, replyPayload)
				if err != nil {
					log.Printf("ERROR: failed to send reply: %v", err)
					cancel()
					if ackErr := d.Ack(false); ackErr != nil {
						return ackErr
					}
					continue
				}
			}
			cancel()
			if err := d.Ack(false); err != nil {
				return err
			}
		}
	}
}

func (b *Broker) replyConfirmed(ctx context.Context, ch replyPublisher, d amqp091.Delivery, payload any) error {
	responseBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, "", d.ReplyTo, true, false, amqp091.Publishing{
		ContentType:   "application/json",
		CorrelationId: d.CorrelationId,
		Body:          responseBytes,
	})
	if err != nil {
		return err
	}

	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return fmt.Errorf("RabbitMQ rejected reply")
	}

	return nil
}
