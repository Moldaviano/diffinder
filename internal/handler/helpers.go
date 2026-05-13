package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// pathUUID estrae un parametro UUID dal path, scrivendo errore 400 se invalido.
func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	v := chi.URLParam(r, name)
	id, err := uuid.Parse(v)
	if err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest("invalid "+name))
		return uuid.Nil, false
	}
	return id, true
}

// decodeJSON helper: decodifica e in caso di errore scrive risposta 400.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest("invalid json: "+err.Error()))
		return false
	}
	return true
}

// readAll: utile per il webhook (raw body per HMAC).
func readAll(r io.Reader) ([]byte, error) { return io.ReadAll(r) }

// principalUserUUID estrae l'UUID dell'utente dal context. Ritorna nil se anonimo.
func principalUserUUID(r *http.Request) *uuid.UUID {
	p, ok := middleware.GetPrincipal(r.Context())
	if !ok {
		return nil
	}
	id, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil
	}
	return &id
}
