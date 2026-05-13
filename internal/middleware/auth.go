package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
)

type ctxKey string

const userCtxKey ctxKey = "auth.user"

// Principal: payload utente messo in context dal middleware Auth.
type Principal struct {
	UserID string
	Email  string
	Role   model.UserRole
}

// Auth middleware: verifica Bearer token e popola il context.
func Auth(issuer *auth.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				httpx.WriteError(w, httpx.ErrUnauthorized("missing bearer token"))
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := issuer.Parse(token)
			if err != nil {
				httpx.WriteError(w, httpx.ErrUnauthorized("invalid token"))
				return
			}
			if claims.Type != auth.TokenAccess {
				httpx.WriteError(w, httpx.ErrUnauthorized("wrong token type"))
				return
			}
			p := Principal{
				UserID: claims.UserID.String(),
				Email:  claims.Email,
				Role:   claims.Role,
			}
			ctx := context.WithValue(r.Context(), userCtxKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetPrincipal estrae il Principal dal context (mai panico, ok=false se mancante).
func GetPrincipal(ctx context.Context) (Principal, bool) {
	v, ok := ctx.Value(userCtxKey).(Principal)
	return v, ok
}

// RequireRole: middleware che richiede uno dei ruoli passati.
func RequireRole(roles ...model.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := GetPrincipal(r.Context())
			if !ok {
				httpx.WriteError(w, httpx.ErrUnauthorized("not authenticated"))
				return
			}
			for _, role := range roles {
				if p.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, httpx.ErrForbidden("insufficient role"))
		})
	}
}
