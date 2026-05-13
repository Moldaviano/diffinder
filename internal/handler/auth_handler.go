package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{svc: s} }

func (h *AuthHandler) Mount(r chi.Router) {
	r.Post("/login", h.login)
	r.Post("/refresh", h.refresh)
	r.Post("/logout", h.logout)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest("invalid json"))
		return
	}
	if in.Email == "" || in.Password == "" {
		httpx.WriteError(w, httpx.ErrBadRequest("email and password are required"))
		return
	}
	tok, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tok)
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		httpx.WriteError(w, httpx.ErrBadRequest("refresh_token is required"))
		return
	}
	tok, err := h.svc.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tok)
}

// logout: con JWT stateless è no-op lato server.
// Il client cancella i token. Mantenuto per coerenza con la spec.
func (h *AuthHandler) logout(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}
