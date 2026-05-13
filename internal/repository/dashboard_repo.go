package repository

import "context"

type DashboardRepo struct{ db *DB }

func NewDashboardRepo(db *DB) *DashboardRepo { return &DashboardRepo{db: db} }

type Summary struct {
	TotalReleases    int `json:"total_releases"`
	InCert           int `json:"in_cert"`
	BlockedPRs       int `json:"blocked_prs"`
	DeploymentsToday int `json:"deployments_today"`
}

func (r *DashboardRepo) Summary(ctx context.Context) (*Summary, error) {
	const q = `
		SELECT
		  (SELECT COUNT(*) FROM releases),
		  (SELECT COUNT(*) FROM releases WHERE status = 'in_cert'),
		  (SELECT COUNT(*) FROM pull_requests WHERE status = 'blocked'),
		  (SELECT COUNT(*) FROM deployment_events WHERE deployed_at >= CURRENT_DATE)`
	s := &Summary{}
	if err := r.db.Pool.QueryRow(ctx, q).Scan(
		&s.TotalReleases, &s.InCert, &s.BlockedPRs, &s.DeploymentsToday,
	); err != nil {
		return nil, err
	}
	return s, nil
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func (r *DashboardRepo) ReleasesByStatus(ctx context.Context) ([]StatusCount, error) {
	const q = `
		SELECT status::TEXT, COUNT(*)::INT
		FROM releases
		GROUP BY status
		ORDER BY status`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StatusCount
	for rows.Next() {
		var s StatusCount
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
