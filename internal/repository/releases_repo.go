package repository

import (
	"context"
	"strings"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type ReleasesRepo struct{ db *DB }

func NewReleasesRepo(db *DB) *ReleasesRepo { return &ReleasesRepo{db: db} }

func (r *ReleasesRepo) Create(ctx context.Context, rel *model.Release) error {
	const q = `
		INSERT INTO releases (project_id, branch_name, title, description, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, q,
		rel.ProjectID, rel.BranchName, rel.Title, rel.Description, rel.Status, rel.CreatedBy,
	).Scan(&rel.ID, &rel.CreatedAt, &rel.UpdatedAt)
}

func (r *ReleasesRepo) Get(ctx context.Context, id uuid.UUID) (*model.Release, error) {
	const q = `
		SELECT id, project_id, branch_name, title, description, status, created_by, created_at, updated_at
		FROM releases WHERE id = $1`
	var rel model.Release
	err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&rel.ID, &rel.ProjectID, &rel.BranchName, &rel.Title, &rel.Description,
		&rel.Status, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// GetByProjectAndBranch usato dal webhook per risalire alla release a partire da repo+branch.
func (r *ReleasesRepo) GetByProjectAndBranch(ctx context.Context, projectID uuid.UUID, branch string) (*model.Release, error) {
	const q = `
		SELECT id, project_id, branch_name, title, description, status, created_by, created_at, updated_at
		FROM releases WHERE project_id = $1 AND branch_name = $2`
	var rel model.Release
	err := r.db.Pool.QueryRow(ctx, q, projectID, branch).Scan(
		&rel.ID, &rel.ProjectID, &rel.BranchName, &rel.Title, &rel.Description,
		&rel.Status, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *ReleasesRepo) Update(ctx context.Context, rel *model.Release) error {
	const q = `
		UPDATE releases
		   SET title = $2, description = $3, status = $4::release_status
		 WHERE id = $1
		 RETURNING updated_at`
	return r.db.Pool.QueryRow(ctx, q, rel.ID, rel.Title, rel.Description, rel.Status).Scan(&rel.UpdatedAt)
}

func (r *ReleasesRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReleaseStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE releases SET status = $2::release_status WHERE id = $1`, id, status)
	return err
}

// ListFilter rappresenta i filtri opzionali per GET /releases.
type ListFilter struct {
	ProjectID *uuid.UUID
	Status    *model.ReleaseStatus
}

func (r *ReleasesRepo) List(ctx context.Context, f ListFilter, limit, offset int) ([]model.Release, int, error) {
	var where []string
	var args []any
	idx := 1
	if f.ProjectID != nil {
		where = append(where, "project_id = $"+itoa(idx))
		args = append(args, *f.ProjectID)
		idx++
	}
	if f.Status != nil {
		where = append(where, "status = $"+itoa(idx))
		args = append(args, *f.Status)
		idx++
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	q := `
		SELECT id, project_id, branch_name, title, description, status, created_by, created_at, updated_at,
		       COUNT(*) OVER()
		FROM releases ` + whereSQL + `
		ORDER BY created_at DESC
		LIMIT $` + itoa(idx) + ` OFFSET $` + itoa(idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.Release
	var total int
	for rows.Next() {
		var rel model.Release
		if err := rows.Scan(&rel.ID, &rel.ProjectID, &rel.BranchName, &rel.Title, &rel.Description,
			&rel.Status, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, rel)
	}
	return out, total, rows.Err()
}

func (r *ReleasesRepo) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Release, int, error) {
	pid := projectID
	return r.List(ctx, ListFilter{ProjectID: &pid}, limit, offset)
}

// itoa minimale per costruzione dinamica delle query (evita strconv per inline).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(b[pos:])
}
