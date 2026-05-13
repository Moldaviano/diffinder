package service

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
)

type DashboardService struct {
	dash        *repository.DashboardRepo
	deployments *repository.DeploymentsRepo
	prs         *repository.PullRequestsRepo
}

func NewDashboardService(
	d *repository.DashboardRepo,
	de *repository.DeploymentsRepo,
	p *repository.PullRequestsRepo,
) *DashboardService {
	return &DashboardService{dash: d, deployments: de, prs: p}
}

func (s *DashboardService) Summary(ctx context.Context) (*repository.Summary, error) {
	return s.dash.Summary(ctx)
}

func (s *DashboardService) ReleasesByStatus(ctx context.Context) ([]repository.StatusCount, error) {
	return s.dash.ReleasesByStatus(ctx)
}

func (s *DashboardService) RecentActivity(ctx context.Context, limit int) ([]repository.ActivityItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.deployments.RecentActivity(ctx, limit)
}

func (s *DashboardService) BlockedPRs(ctx context.Context, limit, offset int) ([]model.PullRequest, int, error) {
	return s.prs.ListBlocked(ctx, limit, offset)
}
