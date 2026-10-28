package rabbitmq

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestBrokerRegisterConsumer(t *testing.T) {
	b := &Broker{}
	handler := func(context.Context, amqp091.Delivery) (any, error) { return nil, nil }
	b.RegisterConsumer("cv.requests", 3, handler)
	if len(b.consumers) != 1 || b.consumers[0].QueueName != "cv.requests" || b.consumers[0].ConsumerCount != 3 || b.consumers[0].Handler == nil {
		t.Fatalf("consumer registration = %+v", b.consumers)
	}
}
