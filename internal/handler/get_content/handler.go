package get_content

import (
	"context"
	"encoding/json"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
)

type ContentProcess interface {
	Execute(lang string) (map[string]interface{}, error)
}

type Handler struct {
	contentv1.UnimplementedContentServiceServer
	process ContentProcess
}

func NewHandler(process ContentProcess) *Handler {
	return &Handler{process: process}
}

func (h *Handler) GetContent(ctx context.Context, req *contentv1.GetContentRequest) (*contentv1.GetContentResponse, error) {
	content, err := h.process.Execute(req.GetLang())
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return &contentv1.GetContentResponse{JsonContent: string(data)}, nil
}
