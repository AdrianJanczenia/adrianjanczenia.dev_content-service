package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

type mockAmqpChannel struct {
	publishedMsg amqp091.Publishing
	publishCtx   context.Context
	routingKey   string
}

func (m *mockAmqpChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	m.publishCtx = ctx
	m.routingKey = key
	m.publishedMsg = msg
	return nil
}

func (m *mockAmqpChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	return nil, nil
}

func (m *mockAmqpChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	return amqp091.Queue{}, nil
}

func (m *mockAmqpChannel) QueueBind(name, key, exchange string, noWait bool, args amqp091.Table) error {
	return nil
}

func (m *mockAmqpChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return nil
}

func (m *mockAmqpChannel) Close() error {
	return nil
}

func TestBroker_Reply(t *testing.T) {
	b := &Broker{}
	ctx := context.Background()
	mockCh := &mockAmqpChannel{}

	delivery := amqp091.Delivery{
		ReplyTo:       "response-queue",
		CorrelationId: "test-corr-id",
	}

	payload := map[string]string{"result": "ok"}

	err := b.reply(ctx, mockCh, delivery, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockCh.routingKey != "response-queue" {
		t.Errorf("expected routing key 'response-queue', got '%s'", mockCh.routingKey)
	}

	if mockCh.publishedMsg.CorrelationId != "test-corr-id" {
		t.Errorf("expected correlation id 'test-corr-id', got '%s'", mockCh.publishedMsg.CorrelationId)
	}

	var sentPayload map[string]string
	json.Unmarshal(mockCh.publishedMsg.Body, &sentPayload)
	if sentPayload["result"] != "ok" {
		t.Errorf("expected payload result 'ok', got '%v'", sentPayload["result"])
	}
}

func TestBroker_RegisterConsumer(t *testing.T) {
	b := &Broker{}
	handler := func(ctx context.Context, d amqp091.Delivery) (any, error) { return nil, nil }

	b.RegisterConsumer("test-queue", 3, handler)

	if len(b.consumers) != 1 {
		t.Fatalf("expected 1 consumer, got %d", len(b.consumers))
	}

	if b.consumers[0].QueueName != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got '%s'", b.consumers[0].QueueName)
	}

	if b.consumers[0].ConsumerCount != 3 {
		t.Errorf("expected consumer count 3, got %d", b.consumers[0].ConsumerCount)
	}
}
