package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type CommitsRepo struct{ db *DB }

func NewCommitsRepo(db *DB) *CommitsRepo { return &CommitsRepo{db: db} }

// BulkUpsert inserisce un set di commit per una release (idempotente sulla coppia release_id+commit_sha).
func (r *CommitsRepo) BulkUpsert(ctx context.Context, releaseID uuid.UUID, commits []model.CommitSnapshot) error {
	if len(commits) == 0 {
		return nil
	}
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO commit_snapshots (release_id, commit_sha, commit_message, author, committed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (release_id, commit_sha) DO NOTHING`
	for _, c := range commits {
		if _, err := tx.Exec(ctx, q, releaseID, c.CommitSHA, c.CommitMessage, c.Author, c.CommittedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CommitsRepo) ListByRelease(ctx context.Context, releaseID uuid.UUID) ([]model.CommitSnapshot, error) {
	const q = `
		SELECT id, release_id, commit_sha, commit_message, author, committed_at, captured_at
		FROM commit_snapshots WHERE release_id = $1
		ORDER BY committed_at DESC`
	rows, err := r.db.Pool.Query(ctx, q, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.CommitSnapshot
	for rows.Next() {
		var c model.CommitSnapshot
		if err := rows.Scan(&c.ID, &c.ReleaseID, &c.CommitSHA, &c.CommitMessage,
			&c.Author, &c.CommittedAt, &c.CapturedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ExistsForRelease verifica se uno specifico commit_sha è stato catturato per quella release.
// Usato dal cert-check come fallback "sha è incluso nello snapshot".
func (r *CommitsRepo) ExistsForRelease(ctx context.Context, releaseID uuid.UUID, sha string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM commit_snapshots WHERE release_id = $1 AND commit_sha = $2)`
	var ok bool
	err := r.db.Pool.QueryRow(ctx, q, releaseID, sha).Scan(&ok)
	return ok, err
}
