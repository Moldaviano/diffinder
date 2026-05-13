package handler

import (
	"net/http"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
)

type PullRequestsHandler struct {
	prs    *service.PRService
	checks *service.CheckService
}

func NewPullRequestsHandler(prs *service.PRService, c *service.CheckService) *PullRequestsHandler {
	return &PullRequestsHandler{prs: prs, checks: c}
}

func (h *PullRequestsHandler) Mount(r chi.Router) {
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.get)
	r.Put("/{id}/status", h.updateStatus)
	r.Post("/{id}/check-cert", h.runCheck)
	r.Get("/{id}/checks", h.checksHistory)
}

func (h *PullRequestsHandler) create(w http.ResponseWriter, r *http.Request) {
	var pr model.PullRequest
	if !decodeJSON(w, r, &pr) {
		return
	}
	if err := h.prs.Create(r.Context(), &pr); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, pr)
}

func (h *PullRequestsHandler) list(w http.ResponseWriter, r *http.Request) {
	pg := httpx.ParsePage(r)
	if r.URL.Query().Get("blocked") == "true" {
		items, total, err := h.prs.ListBlocked(r.Context(), pg.Limit, pg.Offset())
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.PullRequest]{
			Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
		})
		return
	}
	items, total, err := h.prs.List(r.Context(), pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.PullRequest]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}

func (h *PullRequestsHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	pr, err := h.prs.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, pr)
}

type statusReq struct {
	Status string `json:"status"`
}

func (h *PullRequestsHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var req statusReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.prs.UpdateStatus(r.Context(), id, model.PRStatus(req.Status)); err != nil {
		httpx.WriteError(w, err)
		return
	}
	pr, err := h.prs.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, pr)
}

func (h *PullRequestsHandler) runCheck(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	c, err := h.checks.RunCheck(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

func (h *PullRequestsHandler) checksHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.prs.Checks(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}
