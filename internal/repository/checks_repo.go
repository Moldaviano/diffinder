package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type ChecksRepo struct{ db *DB }

func NewChecksRepo(db *DB) *ChecksRepo { return &ChecksRepo{db: db} }

func (r *ChecksRepo) Create(ctx context.Context, c *model.CertificationCheck) error {
	const q = `
		INSERT INTO certification_checks (pull_request_id, head_commit_sha, cert_commit_sha, passed, details)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, checked_at`
	return r.db.Pool.QueryRow(ctx, q,
		c.PullRequestID, c.HeadCommitSHA, c.CertCommitSHA, c.Passed, c.Details,
	).Scan(&c.ID, &c.CheckedAt)
}

func (r *ChecksRepo) ListByPR(ctx context.Context, prID uuid.UUID) ([]model.CertificationCheck, error) {
	const q = `
		SELECT id, pull_request_id, head_commit_sha, cert_commit_sha, passed, checked_at, details
		FROM certification_checks WHERE pull_request_id = $1
		ORDER BY checked_at DESC`
	rows, err := r.db.Pool.Query(ctx, q, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.CertificationCheck
	for rows.Next() {
		var c model.CertificationCheck
		if err := rows.Scan(&c.ID, &c.PullRequestID, &c.HeadCommitSHA, &c.CertCommitSHA,
			&c.Passed, &c.CheckedAt, &c.Details); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// LastByPR ritorna l'ultimo check per una PR (utile per stato semaforo).
func (r *ChecksRepo) LastByPR(ctx context.Context, prID uuid.UUID) (*model.CertificationCheck, error) {
	const q = `
		SELECT id, pull_request_id, head_commit_sha, cert_commit_sha, passed, checked_at, details
		FROM certification_checks WHERE pull_request_id = $1
		ORDER BY checked_at DESC LIMIT 1`
	var c model.CertificationCheck
	err := r.db.Pool.QueryRow(ctx, q, prID).Scan(
		&c.ID, &c.PullRequestID, &c.HeadCommitSHA, &c.CertCommitSHA, &c.Passed, &c.CheckedAt, &c.Details,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
