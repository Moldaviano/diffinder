package service

import (
	"context"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

type PRService struct {
	prs    *repository.PullRequestsRepo
	checks *repository.ChecksRepo
}

func NewPRService(prs *repository.PullRequestsRepo, checks *repository.ChecksRepo) *PRService {
	return &PRService{prs: prs, checks: checks}
}

func (s *PRService) Create(ctx context.Context, pr *model.PullRequest) error {
	if pr.PRURL == "" || pr.HeadCommitSHA == "" || pr.BaseBranch == "" {
		return httpx.ErrBadRequest("pr_url, head_commit_sha, base_branch are required")
	}
	if pr.PRNumber <= 0 {
		return httpx.ErrBadRequest("pr_number must be positive")
	}
	if pr.Status == "" {
		pr.Status = model.PROpen
	}
	if !pr.Status.Valid() {
		return httpx.ErrBadRequest("invalid status")
	}
	return s.prs.Create(ctx, pr)
}

func (s *PRService) Get(ctx context.Context, id uuid.UUID) (*model.PullRequest, error) {
	pr, err := s.prs.Get(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("pull request not found")
		}
		return nil, err
	}
	return pr, nil
}

func (s *PRService) UpdateStatus(ctx context.Context, id uuid.UUID, st model.PRStatus) error {
	if !st.Valid() {
		return httpx.ErrBadRequest("invalid status")
	}
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return s.prs.UpdateStatus(ctx, id, st)
}

func (s *PRService) List(ctx context.Context, limit, offset int) ([]model.PullRequest, int, error) {
	return s.prs.List(ctx, limit, offset)
}

func (s *PRService) ListBlocked(ctx context.Context, limit, offset int) ([]model.PullRequest, int, error) {
	return s.prs.ListBlocked(ctx, limit, offset)
}

func (s *PRService) Checks(ctx context.Context, prID uuid.UUID) ([]model.CertificationCheck, error) {
	return s.checks.ListByPR(ctx, prID)
}
