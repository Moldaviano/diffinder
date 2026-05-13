package service

import (
	"context"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

type ReleaseService struct {
	releases    *repository.ReleasesRepo
	deployments *repository.DeploymentsRepo
	commits     *repository.CommitsRepo
	prs         *repository.PullRequestsRepo
}

func NewReleaseService(
	r *repository.ReleasesRepo,
	d *repository.DeploymentsRepo,
	c *repository.CommitsRepo,
	p *repository.PullRequestsRepo,
) *ReleaseService {
	return &ReleaseService{releases: r, deployments: d, commits: c, prs: p}
}

func (s *ReleaseService) Create(ctx context.Context, rel *model.Release) error {
	if rel.BranchName == "" || rel.Title == "" {
		return httpx.ErrBadRequest("branch_name and title are required")
	}
	if rel.Status == "" {
		rel.Status = model.ReleaseDraft
	}
	if !rel.Status.Valid() {
		return httpx.ErrBadRequest("invalid status")
	}
	return s.releases.Create(ctx, rel)
}

func (s *ReleaseService) Get(ctx context.Context, id uuid.UUID) (*model.Release, error) {
	rel, err := s.releases.Get(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("release not found")
		}
		return nil, err
	}
	return rel, nil
}

func (s *ReleaseService) Update(ctx context.Context, rel *model.Release) error {
	if !rel.Status.Valid() {
		return httpx.ErrBadRequest("invalid status")
	}
	return s.releases.Update(ctx, rel)
}

func (s *ReleaseService) List(ctx context.Context, f repository.ListFilter, limit, offset int) ([]model.Release, int, error) {
	return s.releases.List(ctx, f, limit, offset)
}

func (s *ReleaseService) ListByProject(ctx context.Context, pid uuid.UUID, limit, offset int) ([]model.Release, int, error) {
	return s.releases.ListByProject(ctx, pid, limit, offset)
}

func (s *ReleaseService) Deployments(ctx context.Context, id uuid.UUID) ([]model.DeploymentEvent, error) {
	return s.deployments.ListByRelease(ctx, id)
}

func (s *ReleaseService) PullRequests(ctx context.Context, id uuid.UUID) ([]model.PullRequest, error) {
	return s.prs.ListByRelease(ctx, id)
}

func (s *ReleaseService) Commits(ctx context.Context, id uuid.UUID) ([]model.CommitSnapshot, error) {
	return s.commits.ListByRelease(ctx, id)
}
