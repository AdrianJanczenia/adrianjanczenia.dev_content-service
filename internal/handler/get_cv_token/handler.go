package get_cv_token

import (
	"context"
	"encoding/json"
	"errors"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"github.com/rabbitmq/amqp091-go"
)

type GetCVTokenProcess interface {
	Process(ctx context.Context, password, lang string) (string, error)
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

func (c *Handler) Handle(ctx context.Context, d amqp091.Delivery) (any, error) {
	var req requestPayload
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return nil, appErrors.ErrInvalidInput
	}

	token, err := c.getCVTokenProcess.Process(ctx, req.Password, req.Lang)

	response := responsePayload{}
	if err != nil {
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			response.Error = appErr.Slug
		} else {
			response.Error = appErrors.ErrInternalServerError.Slug
		}
	} else {
		response.Token = token
	}

	return response, nil
}
