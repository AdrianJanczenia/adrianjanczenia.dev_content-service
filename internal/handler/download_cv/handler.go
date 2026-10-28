package download_cv

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type DownloadCVProcess interface {
	Process(ctx context.Context, token, lang string) (io.ReadCloser, int64, error)
}

type Handler struct {
	downloadCVProcess DownloadCVProcess
}

func NewHandler(cvProcess DownloadCVProcess) *Handler {
	return &Handler{downloadCVProcess: cvProcess}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	file, size, err := h.downloadCVProcess.Process(r.Context(), token, lang)
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\"cv_adrian_janczenia.pdf\"")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	_, _ = io.Copy(w, file)
}
