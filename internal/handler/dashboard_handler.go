package handler

import (
	"net/http"
	"strconv"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
)

type DashboardHandler struct{ svc *service.DashboardService }

func NewDashboardHandler(s *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: s}
}

func (h *DashboardHandler) Mount(r chi.Router) {
	r.Get("/summary", h.summary)
	r.Get("/releases-by-status", h.byStatus)
	r.Get("/recent-activity", h.recent)
	r.Get("/blocked-prs", h.blocked)
}

func (h *DashboardHandler) summary(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Summary(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, s)
}

func (h *DashboardHandler) byStatus(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ReleasesByStatus(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *DashboardHandler) recent(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.RecentActivity(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *DashboardHandler) blocked(w http.ResponseWriter, r *http.Request) {
	pg := httpx.ParsePage(r)
	items, total, err := h.svc.BlockedPRs(r.Context(), pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.PullRequest]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}
