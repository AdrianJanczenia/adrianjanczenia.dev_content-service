package download_cv

import (
	"log"
	"net/http"
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
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		http.Error(w, "Bad Request: missing lang parameter", http.StatusBadRequest)
		return
	}

	filePath, err := h.downloadCVProcess.Process(token, lang)
	if err != nil {
		log.Printf("ERROR: CV download failed for token %s: %v", token, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"cv.pdf\"")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, filePath)
}
