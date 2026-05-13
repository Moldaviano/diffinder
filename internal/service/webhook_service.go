package service

import (
	"context"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
)

type WebhookService struct {
	projects *repository.ProjectsRepo
	releases *repository.ReleasesRepo
	prs      *repository.PullRequestsRepo
	check    *CheckService
}

func NewWebhookService(
	p *repository.ProjectsRepo,
	r *repository.ReleasesRepo,
	prs *repository.PullRequestsRepo,
	c *CheckService,
) *WebhookService {
	return &WebhookService{projects: p, releases: r, prs: prs, check: c}
}

// GitHubPRPayload è la forma del JSON che ci aspettiamo da GitHub Actions.
type GitHubPRPayload struct {
	Repo       string `json:"repo"`
	PRNumber   int    `json:"pr_number"`
	HeadSHA    string `json:"head_sha"`
	BaseBranch string `json:"base_branch"`
	PRURL      string `json:"pr_url"`     // opzionale
	HeadBranch string `json:"head_branch"`// opzionale, fallback per identificare la release
}

type WebhookResult struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
}

// HandlePR è il corpo logico del webhook: trova progetto/release/PR,
// crea/aggiorna la PR e lancia il cert-check.
func (s *WebhookService) HandlePR(ctx context.Context, p GitHubPRPayload) (*WebhookResult, error) {
	if p.Repo == "" || p.PRNumber == 0 || p.HeadSHA == "" || p.BaseBranch == "" {
		return nil, httpx.ErrBadRequest("missing required fields")
	}

	project, err := s.projects.GetByRepositoryURL(ctx, p.Repo)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("project not registered for repo " + p.Repo)
		}
		return nil, err
	}

	// Identifica la release dal branch sorgente della PR (head_branch).
	// Se mancante, ricadiamo su base_branch (caso meno preciso).
	branch := p.HeadBranch
	if branch == "" {
		branch = p.BaseBranch
	}
	release, err := s.releases.GetByProjectAndBranch(ctx, project.ID, branch)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("no release tracked for branch " + branch)
		}
		return nil, err
	}

	pr := &model.PullRequest{
		ReleaseID:     release.ID,
		PRURL:         p.PRURL,
		PRNumber:      p.PRNumber,
		HeadCommitSHA: p.HeadSHA,
		BaseBranch:    p.BaseBranch,
		Status:        model.PROpen,
	}
	if err := s.prs.Create(ctx, pr); err != nil {
		return nil, err
	}

	check, err := s.check.RunCheck(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	return &WebhookResult{Passed: check.Passed, Reason: check.Details}, nil
}
