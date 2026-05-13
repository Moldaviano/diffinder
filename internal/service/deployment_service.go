package service

import (
	"context"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

type DeploymentService struct {
	releases    *repository.ReleasesRepo
	deployments *repository.DeploymentsRepo
	commits     *repository.CommitsRepo
}

func NewDeploymentService(
	r *repository.ReleasesRepo,
	d *repository.DeploymentsRepo,
	c *repository.CommitsRepo,
) *DeploymentService {
	return &DeploymentService{releases: r, deployments: d, commits: c}
}

// DeployInput è il payload accettato per registrare un deploy.
// Se environment=="cert" e Commits!=nil, popoliamo il commit_snapshots.
type DeployInput struct {
	ReleaseID   uuid.UUID
	Environment model.Environment
	CommitSHA   string
	DeployedBy  *uuid.UUID
	Notes       string
	Commits     []model.CommitSnapshot // opzionali, attesi quando environment=cert
}

// Register valida e registra l'evento. Aggiorna lo status della release
// in base all'ambiente (dev → in_dev, cert → in_cert, prod → in_prod).
// Quando si deploya in cert, persiste lo snapshot dei commit forniti.
func (s *DeploymentService) Register(ctx context.Context, in DeployInput) (*model.DeploymentEvent, error) {
	if !in.Environment.Valid() {
		return nil, httpx.ErrBadRequest("invalid environment")
	}
	if in.CommitSHA == "" {
		return nil, httpx.ErrBadRequest("commit_sha is required")
	}
	rel, err := s.releases.Get(ctx, in.ReleaseID)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("release not found")
		}
		return nil, err
	}

	ev := &model.DeploymentEvent{
		ReleaseID:   in.ReleaseID,
		Environment: in.Environment,
		CommitSHA:   in.CommitSHA,
		DeployedBy:  in.DeployedBy,
		Notes:       in.Notes,
	}
	if err := s.deployments.Create(ctx, ev); err != nil {
		return nil, err
	}

	// Update release status in base all'ambiente
	var newStatus model.ReleaseStatus
	switch in.Environment {
	case model.EnvDev:
		newStatus = model.ReleaseInDev
	case model.EnvCert:
		newStatus = model.ReleaseInCert
		if len(in.Commits) > 0 {
			if err := s.commits.BulkUpsert(ctx, in.ReleaseID, in.Commits); err != nil {
				return nil, err
			}
		}
	case model.EnvProd:
		newStatus = model.ReleaseInProd
	}
	if newStatus != "" && rel.Status != newStatus {
		if err := s.releases.UpdateStatus(ctx, in.ReleaseID, newStatus); err != nil {
			return nil, err
		}
	}
	return ev, nil
}
