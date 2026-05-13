package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// AppError è l'errore di dominio che gli handler propagano verso il client.
type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"error"`
}

func (e *AppError) Error() string { return e.Message }

func NewError(status int, code, msg string) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: msg}
}

// Codici errore standard
var (
	ErrBadRequest   = func(msg string) *AppError { return NewError(http.StatusBadRequest, "BAD_REQUEST", msg) }
	ErrUnauthorized = func(msg string) *AppError { return NewError(http.StatusUnauthorized, "UNAUTHORIZED", msg) }
	ErrForbidden    = func(msg string) *AppError { return NewError(http.StatusForbidden, "FORBIDDEN", msg) }
	ErrNotFound     = func(msg string) *AppError { return NewError(http.StatusNotFound, "NOT_FOUND", msg) }
	ErrConflict     = func(msg string) *AppError { return NewError(http.StatusConflict, "CONFLICT", msg) }
	ErrInternal     = func(msg string) *AppError { return NewError(http.StatusInternalServerError, "INTERNAL", msg) }
)

// WriteJSON scrive una risposta JSON con lo status fornito.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

// WriteError converte un errore in risposta JSON uniforme.
// Se l'errore è un *AppError usa code/status/message; altrimenti 500 INTERNAL.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		WriteJSON(w, appErr.HTTPStatus, appErr)
		return
	}
	WriteJSON(w, http.StatusInternalServerError, &AppError{
		Code: "INTERNAL", Message: "internal server error",
	})
}
