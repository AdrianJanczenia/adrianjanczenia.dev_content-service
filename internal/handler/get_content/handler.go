package get_content

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
)

type GetContentProcess interface {
	Process(lang string) ([]byte, error)
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

	return &contentv1.GetContentResponse{JsonContent: content}, nil
}
