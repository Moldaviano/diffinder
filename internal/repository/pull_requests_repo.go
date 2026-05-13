package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type PullRequestsRepo struct{ db *DB }

func NewPullRequestsRepo(db *DB) *PullRequestsRepo { return &PullRequestsRepo{db: db} }

func (r *PullRequestsRepo) Create(ctx context.Context, pr *model.PullRequest) error {
	const q = `
		INSERT INTO pull_requests (release_id, pr_url, pr_number, head_commit_sha, base_branch, status, opened_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, NOW()), $8)
		ON CONFLICT (release_id, pr_number) DO UPDATE
		   SET head_commit_sha = EXCLUDED.head_commit_sha,
		       base_branch     = EXCLUDED.base_branch,
		       status          = EXCLUDED.status
		RETURNING id, opened_at`
	var openedAt any
	if !pr.OpenedAt.IsZero() {
		openedAt = pr.OpenedAt
	}
	return r.db.Pool.QueryRow(ctx, q,
		pr.ReleaseID, pr.PRURL, pr.PRNumber, pr.HeadCommitSHA, pr.BaseBranch, pr.Status, openedAt, pr.MergedAt,
	).Scan(&pr.ID, &pr.OpenedAt)
}

func (r *PullRequestsRepo) Get(ctx context.Context, id uuid.UUID) (*model.PullRequest, error) {
	const q = `
		SELECT id, release_id, pr_url, pr_number, head_commit_sha, base_branch, status, opened_at, merged_at
		FROM pull_requests WHERE id = $1`
	var pr model.PullRequest
	err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&pr.ID, &pr.ReleaseID, &pr.PRURL, &pr.PRNumber, &pr.HeadCommitSHA,
		&pr.BaseBranch, &pr.Status, &pr.OpenedAt, &pr.MergedAt,
	)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PullRequestsRepo) GetByReleaseAndNumber(ctx context.Context, releaseID uuid.UUID, number int) (*model.PullRequest, error) {
	const q = `
		SELECT id, release_id, pr_url, pr_number, head_commit_sha, base_branch, status, opened_at, merged_at
		FROM pull_requests WHERE release_id = $1 AND pr_number = $2`
	var pr model.PullRequest
	err := r.db.Pool.QueryRow(ctx, q, releaseID, number).Scan(
		&pr.ID, &pr.ReleaseID, &pr.PRURL, &pr.PRNumber, &pr.HeadCommitSHA,
		&pr.BaseBranch, &pr.Status, &pr.OpenedAt, &pr.MergedAt,
	)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PullRequestsRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PRStatus) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE pull_requests
		   SET status = $2,
		       merged_at = CASE WHEN $2 = 'merged' AND merged_at IS NULL THEN NOW() ELSE merged_at END
		 WHERE id = $1`, id, status)
	return err
}

func (r *PullRequestsRepo) ListByRelease(ctx context.Context, releaseID uuid.UUID) ([]model.PullRequest, error) {
	const q = `
		SELECT id, release_id, pr_url, pr_number, head_commit_sha, base_branch, status, opened_at, merged_at
		FROM pull_requests WHERE release_id = $1
		ORDER BY opened_at DESC`
	rows, err := r.db.Pool.Query(ctx, q, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		if err := rows.Scan(&pr.ID, &pr.ReleaseID, &pr.PRURL, &pr.PRNumber, &pr.HeadCommitSHA,
			&pr.BaseBranch, &pr.Status, &pr.OpenedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		out = append(out, pr)
	}
	return out, rows.Err()
}

// ListBlocked: PR il cui ultimo cert-check ha passed=false.
func (r *PullRequestsRepo) ListBlocked(ctx context.Context, limit, offset int) ([]model.PullRequest, int, error) {
	const q = `
		WITH last_check AS (
		  SELECT DISTINCT ON (pull_request_id)
		    pull_request_id, passed, checked_at
		  FROM certification_checks
		  ORDER BY pull_request_id, checked_at DESC
		)
		SELECT pr.id, pr.release_id, pr.pr_url, pr.pr_number, pr.head_commit_sha,
		       pr.base_branch, pr.status, pr.opened_at, pr.merged_at,
		       COUNT(*) OVER()
		FROM pull_requests pr
		JOIN last_check lc ON lc.pull_request_id = pr.id
		WHERE lc.passed = FALSE AND pr.status IN ('open','blocked')
		ORDER BY pr.opened_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.PullRequest
	var total int
	for rows.Next() {
		var pr model.PullRequest
		if err := rows.Scan(&pr.ID, &pr.ReleaseID, &pr.PRURL, &pr.PRNumber, &pr.HeadCommitSHA,
			&pr.BaseBranch, &pr.Status, &pr.OpenedAt, &pr.MergedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, pr)
	}
	return out, total, rows.Err()
}

func (r *PullRequestsRepo) List(ctx context.Context, limit, offset int) ([]model.PullRequest, int, error) {
	const q = `
		SELECT id, release_id, pr_url, pr_number, head_commit_sha, base_branch, status, opened_at, merged_at,
		       COUNT(*) OVER()
		FROM pull_requests
		ORDER BY opened_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.PullRequest
	var total int
	for rows.Next() {
		var pr model.PullRequest
		if err := rows.Scan(&pr.ID, &pr.ReleaseID, &pr.PRURL, &pr.PRNumber, &pr.HeadCommitSHA,
			&pr.BaseBranch, &pr.Status, &pr.OpenedAt, &pr.MergedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, pr)
	}
	return out, total, rows.Err()
}
