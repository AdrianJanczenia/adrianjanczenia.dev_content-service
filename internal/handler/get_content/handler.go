package get_content

import (
	"context"
	"encoding/json"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
)

type GetContentProcess interface {
	Process(lang string) (map[string]interface{}, error)
}

type Handler struct {
	contentv1.UnimplementedContentServiceServer
	getContentProcess GetContentProcess
}

func NewHandler(process GetContentProcess) *Handler {
	return &Handler{getContentProcess: process}
}

func (h *Handler) Handle(ctx context.Context, req *contentv1.GetContentRequest) (*contentv1.GetContentResponse, error) {
	content, err := h.getContentProcess.Process(req.GetLang())
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return &contentv1.GetContentResponse{JsonContent: string(data)}, nil
}
