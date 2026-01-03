package get_cv_token

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type GetCVTokenProcess interface {
	Process(password, lang string) (string, error)
}

type Handler struct {
	getCVTokenProcess GetCVTokenProcess
}

type requestPayload struct {
	Password string `json:"password"`
	Lang     string `json:"lang"`
}

type responsePayload struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewHandler(cvProcess GetCVTokenProcess) *Handler {
	return &Handler{
		getCVTokenProcess: cvProcess,
	}
}

func (c *Handler) Handle(d amqp091.Delivery) (any, error) {
	var req requestPayload
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return nil, err
	}

	token, err := c.getCVTokenProcess.Process(req.Password, req.Lang)

	response := responsePayload{}
	if err != nil {
		response.Error = err.Error()
	} else {
		response.Token = token
	}

	return response, nil
}
