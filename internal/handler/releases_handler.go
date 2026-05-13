package handler

import (
	"net/http"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReleasesHandler struct {
	releases    *service.ReleaseService
	deployments *service.DeploymentService
}

func NewReleasesHandler(r *service.ReleaseService, d *service.DeploymentService) *ReleasesHandler {
	return &ReleasesHandler{releases: r, deployments: d}
}

func (h *ReleasesHandler) Mount(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Get("/{id}/deployments", h.deployments_)
	r.Get("/{id}/pull-requests", h.pullRequests)
	r.Get("/{id}/commits", h.commits)
	r.Post("/{id}/deploy", h.deploy)
}

func (h *ReleasesHandler) list(w http.ResponseWriter, r *http.Request) {
	pg := httpx.ParsePage(r)
	f := repository.ListFilter{}
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err != nil {
			httpx.WriteError(w, httpx.ErrBadRequest("invalid project_id"))
			return
		}
		f.ProjectID = &id
	}
	if st := r.URL.Query().Get("status"); st != "" {
		s := model.ReleaseStatus(st)
		if !s.Valid() {
			httpx.WriteError(w, httpx.ErrBadRequest("invalid status"))
			return
		}
		f.Status = &s
	}
	items, total, err := h.releases.List(r.Context(), f, pg.Limit, pg.Offset())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.PagedResponse[model.Release]{
		Items: items, Page: pg.Page, Limit: pg.Limit, Total: total,
	})
}

func (h *ReleasesHandler) create(w http.ResponseWriter, r *http.Request) {
	var rel model.Release
	if !decodeJSON(w, r, &rel) {
		return
	}
	rel.CreatedBy = principalUserUUID(r)
	if err := h.releases.Create(r.Context(), &rel); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, rel)
}

func (h *ReleasesHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	rel, err := h.releases.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rel)
}

func (h *ReleasesHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	rel, err := h.releases.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if !decodeJSON(w, r, rel) {
		return
	}
	rel.ID = id
	if err := h.releases.Update(r.Context(), rel); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rel)
}

func (h *ReleasesHandler) deployments_(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.releases.Deployments(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *ReleasesHandler) pullRequests(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.releases.PullRequests(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *ReleasesHandler) commits(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.releases.Commits(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// POST /releases/:id/deploy
type deployReq struct {
	Environment string                  `json:"environment"`
	CommitSHA   string                  `json:"commit_sha"`
	Notes       string                  `json:"notes"`
	Commits     []model.CommitSnapshot  `json:"commits,omitempty"` // attesi quando environment=cert
}

func (h *ReleasesHandler) deploy(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var req deployReq
	if !decodeJSON(w, r, &req) {
		return
	}
	ev, err := h.deployments.Register(r.Context(), service.DeployInput{
		ReleaseID:   id,
		Environment: model.Environment(req.Environment),
		CommitSHA:   req.CommitSHA,
		DeployedBy:  principalUserUUID(r),
		Notes:       req.Notes,
		Commits:     req.Commits,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ev)
}
