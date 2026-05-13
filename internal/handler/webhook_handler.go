package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
)

type WebhookHandler struct {
	svc    *service.WebhookService
	secret string
}

func NewWebhookHandler(s *service.WebhookService, secret string) *WebhookHandler {
	return &WebhookHandler{svc: s, secret: secret}
}

func (h *WebhookHandler) Mount(r chi.Router) {
	r.Post("/github/pr", h.githubPR)
}

// verifySig calcola HMAC-SHA256 sul body raw e confronta con il
// valore inviato da GitHub Actions in `X-Hub-Signature-256: sha256=...`.
// Confronto in tempo costante per evitare timing attacks.
func (h *WebhookHandler) verifySig(body []byte, sigHeader string) bool {
	if !strings.HasPrefix(sigHeader, "sha256=") {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(sigHeader, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	want := mac.Sum(nil)
	return hmac.Equal(got, want)
}

func (h *WebhookHandler) githubPR(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest("cannot read body"))
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySig(body, sig) {
		httpx.WriteError(w, httpx.ErrUnauthorized("invalid signature"))
		return
	}

	var p service.GitHubPRPayload
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&p); err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest("invalid json: "+err.Error()))
		return
	}

	res, err := h.svc.HandlePR(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}
