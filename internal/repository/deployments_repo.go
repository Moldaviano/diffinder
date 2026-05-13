package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type DeploymentsRepo struct{ db *DB }

func NewDeploymentsRepo(db *DB) *DeploymentsRepo { return &DeploymentsRepo{db: db} }

func (r *DeploymentsRepo) Create(ctx context.Context, d *model.DeploymentEvent) error {
	const q = `
		INSERT INTO deployment_events (release_id, environment, commit_sha, deployed_by, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, deployed_at`
	return r.db.Pool.QueryRow(ctx, q,
		d.ReleaseID, d.Environment, d.CommitSHA, d.DeployedBy, d.Notes,
	).Scan(&d.ID, &d.DeployedAt)
}

func (r *DeploymentsRepo) ListByRelease(ctx context.Context, releaseID uuid.UUID) ([]model.DeploymentEvent, error) {
	const q = `
		SELECT id, release_id, environment, commit_sha, deployed_by, deployed_at, notes
		FROM deployment_events WHERE release_id = $1
		ORDER BY deployed_at ASC`
	rows, err := r.db.Pool.Query(ctx, q, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.DeploymentEvent
	for rows.Next() {
		var d model.DeploymentEvent
		if err := rows.Scan(&d.ID, &d.ReleaseID, &d.Environment, &d.CommitSHA,
			&d.DeployedBy, &d.DeployedAt, &d.Notes); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// LatestCertDeployment ritorna il deploy in cert più recente per una release.
// Usato dal cert-check per ricavare il "cert HEAD".
func (r *DeploymentsRepo) LatestCertDeployment(ctx context.Context, releaseID uuid.UUID) (*model.DeploymentEvent, error) {
	const q = `
		SELECT id, release_id, environment, commit_sha, deployed_by, deployed_at, notes
		FROM deployment_events
		WHERE release_id = $1 AND environment = 'cert'
		ORDER BY deployed_at DESC
		LIMIT 1`
	var d model.DeploymentEvent
	err := r.db.Pool.QueryRow(ctx, q, releaseID).Scan(
		&d.ID, &d.ReleaseID, &d.Environment, &d.CommitSHA, &d.DeployedBy, &d.DeployedAt, &d.Notes,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeploymentsRepo) CountToday(ctx context.Context) (int, error) {
	const q = `SELECT COUNT(*) FROM deployment_events WHERE deployed_at >= CURRENT_DATE`
	var n int
	err := r.db.Pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}

// RecentActivity per la dashboard.
type ActivityItem struct {
	Type        string    `json:"type"` // "deployment" | "pr" | "check"
	ReleaseID   uuid.UUID `json:"release_id"`
	Title       string    `json:"title"`
	Environment string    `json:"environment,omitempty"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
	At          string    `json:"at"`
}

func (r *DeploymentsRepo) RecentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	const q = `
		SELECT 'deployment'::TEXT, de.release_id, r.title, de.environment::TEXT, de.commit_sha, de.deployed_at::TEXT
		FROM deployment_events de
		JOIN releases r ON r.id = de.release_id
		ORDER BY de.deployed_at DESC
		LIMIT $1`
	rows, err := r.db.Pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ActivityItem
	for rows.Next() {
		var a ActivityItem
		if err := rows.Scan(&a.Type, &a.ReleaseID, &a.Title, &a.Environment, &a.CommitSHA, &a.At); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
