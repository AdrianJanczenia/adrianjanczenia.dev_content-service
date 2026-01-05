package download_cv

import (
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type DownloadCVProcess interface {
	Process(token, lang string) (string, error)
}

type Handler struct {
	downloadCVProcess DownloadCVProcess
}

func NewHandler(cvProcess DownloadCVProcess) *Handler {
	return &Handler{downloadCVProcess: cvProcess}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		errors.WriteJSON(w, errors.ErrCVExpired)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	filePath, err := h.downloadCVProcess.Process(token, lang)
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"cv.pdf\"")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, filePath)
}
