package get_cv_link

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type GetCVLinkProcess interface {
	Process(password string) (string, error)
}

type Handler struct {
	getCVLinkProcess GetCVLinkProcess
}

type requestPayload struct {
	Password string `json:"password"`
}

type responsePayload struct {
	URL   string `json:"url,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewHandler(cvProcess GetCVLinkProcess) *Handler {
	return &Handler{
		getCVLinkProcess: cvProcess,
	}
}

func (c *Handler) Handle(d amqp091.Delivery) (any, error) {
	var req requestPayload
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return nil, err
	}

	url, err := c.getCVLinkProcess.Process(req.Password)

	response := responsePayload{}
	if err != nil {
		response.Error = err.Error()
	} else {
		response.URL = url
	}

	return response, nil
}
