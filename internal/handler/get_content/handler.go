package get_content

import (
	"context"
	"errors"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetContentProcess interface {
	Process(ctx context.Context, lang string) ([]byte, error)
}

type Handler struct {
	contentv1.UnimplementedContentServiceServer
	getContentProcess GetContentProcess
}

func NewHandler(process GetContentProcess) *Handler {
	return &Handler{getContentProcess: process}
}

func (h *Handler) Handle(ctx context.Context, req *contentv1.GetContentRequest) (*contentv1.GetContentResponse, error) {
	content, err := h.getContentProcess.Process(ctx, req.GetLang())
	if err != nil {
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			if errors.Is(appErr, appErrors.ErrContentNotFound) {
				return nil, status.Error(codes.NotFound, appErr.Slug)
			}
		}
		return nil, status.Error(codes.Internal, appErrors.ErrInternalServerError.Slug)
	}

	return &contentv1.GetContentResponse{JsonContent: content}, nil
}
