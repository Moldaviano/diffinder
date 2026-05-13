package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type ProjectsRepo struct{ db *DB }

func NewProjectsRepo(db *DB) *ProjectsRepo { return &ProjectsRepo{db: db} }

func (r *ProjectsRepo) Create(ctx context.Context, p *model.Project) error {
	const q = `
		INSERT INTO projects (name, description, repository_url, webhook_token)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, q,
		p.Name, p.Description, p.RepositoryURL, p.WebhookToken,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *ProjectsRepo) Get(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	const q = `
		SELECT id, name, description, repository_url, webhook_token, created_at, updated_at
		FROM projects WHERE id = $1`
	var p model.Project
	err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.RepositoryURL, &p.WebhookToken, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectsRepo) GetByRepositoryURL(ctx context.Context, url string) (*model.Project, error) {
	const q = `
		SELECT id, name, description, repository_url, webhook_token, created_at, updated_at
		FROM projects WHERE repository_url = $1`
	var p model.Project
	err := r.db.Pool.QueryRow(ctx, q, url).Scan(
		&p.ID, &p.Name, &p.Description, &p.RepositoryURL, &p.WebhookToken, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectsRepo) Update(ctx context.Context, p *model.Project) error {
	const q = `
		UPDATE projects
		   SET name = $2, description = $3, repository_url = $4, webhook_token = $5
		 WHERE id = $1
		 RETURNING updated_at`
	return r.db.Pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Description, p.RepositoryURL, p.WebhookToken,
	).Scan(&p.UpdatedAt)
}

func (r *ProjectsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

func (r *ProjectsRepo) List(ctx context.Context, limit, offset int) ([]model.Project, int, error) {
	const q = `
		SELECT id, name, description, repository_url, webhook_token, created_at, updated_at,
		       COUNT(*) OVER()
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.Project
	var total int
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.RepositoryURL, &p.WebhookToken,
			&p.CreatedAt, &p.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

// Stats per progetto: # release attive, ultima attività.
type ProjectStats struct {
	ProjectID      uuid.UUID `json:"project_id"`
	ActiveReleases int       `json:"active_releases"`
	LastActivity   *string   `json:"last_activity,omitempty"` // ISO timestamp, null se nessuna attività
}

func (r *ProjectsRepo) Stats(ctx context.Context, id uuid.UUID) (*ProjectStats, error) {
	const q = `
		SELECT
		  (SELECT COUNT(*) FROM releases WHERE project_id = $1 AND status NOT IN ('in_prod','rejected')) AS active_releases,
		  (SELECT MAX(updated_at)::TEXT FROM releases WHERE project_id = $1) AS last_activity`
	s := &ProjectStats{ProjectID: id}
	if err := r.db.Pool.QueryRow(ctx, q, id).Scan(&s.ActiveReleases, &s.LastActivity); err != nil {
		return nil, err
	}
	return s, nil
}
