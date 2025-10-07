package get_cv_link

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type CVProcess interface {
	GenerateLink(password string) (string, error)
}

type RabbitMQBroker interface {
	Reply(d amqp091.Delivery, payload any) error
}

type Consumer struct {
	cvProcess CVProcess
	broker    RabbitMQBroker
}

type requestPayload struct {
	Password string `json:"password"`
}

type responsePayload struct {
	URL   string `json:"url,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewConsumer(cvProcess CVProcess, broker RabbitMQBroker) *Consumer {
	return &Consumer{
		cvProcess: cvProcess,
		broker:    broker,
	}
}

func (c *Consumer) Handle(d amqp091.Delivery) error {
	var req requestPayload
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return err
	}

	url, err := c.cvProcess.GenerateLink(req.Password)

	response := responsePayload{}
	if err != nil {
		response.Error = err.Error()
	} else {
		response.URL = url
	}

	return c.broker.Reply(d, response)
}
