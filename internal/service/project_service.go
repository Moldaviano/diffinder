package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

type ProjectService struct{ repo *repository.ProjectsRepo }

func NewProjectService(repo *repository.ProjectsRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, p *model.Project) error {
	if p.Name == "" {
		return httpx.ErrBadRequest("name is required")
	}
	if p.WebhookToken == "" {
		p.WebhookToken = randomToken(24)
	}
	return s.repo.Create(ctx, p)
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("project not found")
		}
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Update(ctx context.Context, p *model.Project) error {
	if p.Name == "" {
		return httpx.ErrBadRequest("name is required")
	}
	return s.repo.Update(ctx, p)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProjectService) List(ctx context.Context, limit, offset int) ([]model.Project, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ProjectService) Stats(ctx context.Context, id uuid.UUID) (*repository.ProjectStats, error) {
	return s.repo.Stats(ctx, id)
}

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "" // estremamente improbabile
	}
	return hex.EncodeToString(b)
}
