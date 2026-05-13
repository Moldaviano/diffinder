package handler

import (
	"net/http"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/go-chi/chi/v5"
)

type UsersHandler struct{ repo *repository.UsersRepo }

func NewUsersHandler(r *repository.UsersRepo) *UsersHandler {
	return &UsersHandler{repo: r}
}

func (h *UsersHandler) Mount(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
}

func (h *UsersHandler) list(w http.ResponseWriter, r *http.Request) {
	pg := httpx.ParsePage(r)
	items, total, err := h.repo.List(r.Context(), pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.User]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}

type createUserReq struct {
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Password string         `json:"password"`
	Role     model.UserRole `json:"role"`
}

func (h *UsersHandler) create(w http.ResponseWriter, r *http.Request) {
	var in createUserReq
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Username == "" || in.Email == "" || in.Password == "" {
		httpx.WriteError(w, httpx.ErrBadRequest("username, email, password are required"))
		return
	}
	if in.Role == "" {
		in.Role = model.RoleDeveloper
	}
	if !in.Role.Valid() {
		httpx.WriteError(w, httpx.ErrBadRequest("invalid role"))
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	u := &model.User{
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: hash,
		Role:         in.Role,
	}
	if err := h.repo.Create(r.Context(), u); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}
