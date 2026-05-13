package service

import (
	"context"
	"fmt"

	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

type CheckService struct {
	prs         *repository.PullRequestsRepo
	deployments *repository.DeploymentsRepo
	commits     *repository.CommitsRepo
	checks      *repository.ChecksRepo
}

func NewCheckService(
	prs *repository.PullRequestsRepo,
	d *repository.DeploymentsRepo,
	c *repository.CommitsRepo,
	checks *repository.ChecksRepo,
) *CheckService {
	return &CheckService{prs: prs, deployments: d, commits: c, checks: checks}
}

// RunCheck esegue il controllo per la PR indicata e salva il risultato.
// Regola:
//   - se non esiste un deploy in cert per la release → passed=false, reason="no cert deployment"
//   - se head_sha == cert_head_sha                  → passed=true
//   - se head_sha è presente nello snapshot della release → passed=true (è stato testato)
//   - altrimenti                                          → passed=false, reason="head not certified"
//
// Aggiorna anche lo stato della PR a `blocked` se passed=false (e non era già merged/closed).
func (s *CheckService) RunCheck(ctx context.Context, prID uuid.UUID) (*model.CertificationCheck, error) {
	pr, err := s.prs.Get(ctx, prID)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrNotFound("pull request not found")
		}
		return nil, err
	}

	certDep, err := s.deployments.LatestCertDeployment(ctx, pr.ReleaseID)
	if err != nil && !repository.IsNotFound(err) {
		return nil, err
	}

	check := &model.CertificationCheck{
		PullRequestID: pr.ID,
		HeadCommitSHA: pr.HeadCommitSHA,
	}

	if certDep == nil {
		check.Passed = false
		check.Details = "no cert deployment found for this release"
	} else {
		check.CertCommitSHA = certDep.CommitSHA
		switch {
		case pr.HeadCommitSHA == certDep.CommitSHA:
			check.Passed = true
			check.Details = "head matches cert HEAD"
		default:
			ok, err := s.commits.ExistsForRelease(ctx, pr.ReleaseID, pr.HeadCommitSHA)
			if err != nil {
				return nil, err
			}
			if ok {
				check.Passed = true
				check.Details = "head commit is included in cert snapshot"
			} else {
				check.Passed = false
				check.Details = fmt.Sprintf(
					"head %s not present in cert snapshot (cert HEAD: %s) — possible un-certified commits",
					pr.HeadCommitSHA, certDep.CommitSHA,
				)
			}
		}
	}

	if err := s.checks.Create(ctx, check); err != nil {
		return nil, err
	}

	// Aggiorna lo stato della PR di conseguenza (solo se non terminale).
	if pr.Status == model.PROpen || pr.Status == model.PRBlocked {
		want := model.PROpen
		if !check.Passed {
			want = model.PRBlocked
		}
		if pr.Status != want {
			if err := s.prs.UpdateStatus(ctx, pr.ID, want); err != nil {
				return nil, err
			}
		}
	}
	return check, nil
}
