package handler

import (
	"net/http"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
)

type ProjectsHandler struct {
	projects *service.ProjectService
	releases *service.ReleaseService
}

func NewProjectsHandler(p *service.ProjectService, r *service.ReleaseService) *ProjectsHandler {
	return &ProjectsHandler{projects: p, releases: r}
}

func (h *ProjectsHandler) Mount(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Get("/{id}/releases", h.releasesForProject)
	r.Get("/{id}/stats", h.stats)
}

func (h *ProjectsHandler) list(w http.ResponseWriter, r *http.Request) {
	pg := httpx.ParsePage(r)
	items, total, err := h.projects.List(r.Context(), pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.Project]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}

func (h *ProjectsHandler) create(w http.ResponseWriter, r *http.Request) {
	var p model.Project
	if !decodeJSON(w, r, &p) {
		return
	}
	if err := h.projects.Create(r.Context(), &p); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *ProjectsHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	p, err := h.projects.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *ProjectsHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	p, err := h.projects.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if !decodeJSON(w, r, p) {
		return
	}
	p.ID = id
	if err := h.projects.Update(r.Context(), p); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *ProjectsHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.projects.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectsHandler) releasesForProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	pg := httpx.ParsePage(r)
	items, total, err := h.releases.ListByProject(r.Context(), id, pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.Release]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}

func (h *ProjectsHandler) stats(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	s, err := h.projects.Stats(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, s)
}
