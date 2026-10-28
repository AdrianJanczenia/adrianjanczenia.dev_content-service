package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

type AppError struct {
	HTTPStatus int
	Slug       string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Slug
}

func (e *AppError) Unwrap() error {
	return e.Err
}

var (
	ErrInternalServerError = &AppError{HTTPStatus: http.StatusInternalServerError, Slug: "error_internal"}
	ErrCVServer            = &AppError{HTTPStatus: http.StatusInternalServerError, Slug: "error_cv_server"}
	ErrServiceUnavailable  = &AppError{HTTPStatus: http.StatusServiceUnavailable, Slug: "error_service_unavailable"}
	ErrMethodNotAllowed    = &AppError{HTTPStatus: http.StatusMethodNotAllowed, Slug: "error_method_not_allowed"}
	ErrInvalidInput        = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_invalid_input"}
	ErrUnsupportedLanguage = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_unsupported_language"}
	ErrInvalidPassword     = &AppError{HTTPStatus: http.StatusUnauthorized, Slug: "error_cv_auth"}
	ErrCVNotFound          = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_cv_not_found"}
	ErrCVExpired           = &AppError{HTTPStatus: http.StatusGone, Slug: "error_cv_expired"}
	ErrContentNotFound     = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_content_not_found"}
	ErrCaptchaNotFound     = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_captcha_not_found"}
	ErrCaptchaNotSolved    = &AppError{HTTPStatus: http.StatusForbidden, Slug: "error_captcha_not_solved"}
	ErrNoTriesLeft         = &AppError{HTTPStatus: http.StatusGone, Slug: "error_captcha_expired"}
)

func FromSlug(slug string) *AppError {
	switch slug {
	case "error_cv_auth":
		return ErrInvalidPassword
	case "error_cv_expired":
		return ErrCVExpired
	case "error_cv_not_found":
		return ErrCVNotFound
	case "error_cv_server":
		return ErrCVServer
	case "error_internal":
		return ErrInternalServerError
	case "error_message":
		return ErrServiceUnavailable
	case "error_service_unavailable":
		return ErrServiceUnavailable
	case "error_method_not_allowed":
		return ErrMethodNotAllowed
	case "error_invalid_input":
		return ErrInvalidInput
	case "error_unsupported_language":
		return ErrUnsupportedLanguage
	case "error_content_not_found":
		return ErrContentNotFound
	case "error_captcha_not_found":
		return ErrCaptchaNotFound
	case "error_captcha_not_solved":
		return ErrCaptchaNotSolved
	case "error_captcha_expired":
		return ErrNoTriesLeft
	default:
		return ErrInternalServerError
	}
}

func WriteJSON(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = ErrInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]string{
		"error": appErr.Slug,
	})
}
