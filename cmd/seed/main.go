package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/config"
	"github.com/alloy/diffinder/internal/logger"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	ctx := context.Background()
	db, err := repository.NewDB(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := run(ctx, db, cfg); err != nil {
		log.Error("seed failed", "err", err)
		os.Exit(1)
	}
	log.Info("seed completed successfully")
}

func run(ctx context.Context, db *repository.DB, cfg *config.Config) error {
	usersRepo := repository.NewUsersRepo(db)
	projectsRepo := repository.NewProjectsRepo(db)
	releasesRepo := repository.NewReleasesRepo(db)
	deploymentsRepo := repository.NewDeploymentsRepo(db)
	commitsRepo := repository.NewCommitsRepo(db)
	prsRepo := repository.NewPullRequestsRepo(db)
	checksRepo := repository.NewChecksRepo(db)

	// ----- Users -----
	adminEmail := getenv("SEED_ADMIN_EMAIL", "admin@diffinder.local")
	adminPwd := getenv("SEED_ADMIN_PASSWORD", "admin123")

	users := []struct {
		Username, Email, Pwd string
		Role                 model.UserRole
	}{
		{"admin", adminEmail, adminPwd, model.RoleAdmin},
		{"alice", "alice@diffinder.local", "alice123", model.RoleDeveloper},
		{"bob", "bob@diffinder.local", "bob123", model.RoleDeveloper},
		{"viewer", "viewer@diffinder.local", "viewer123", model.RoleViewer},
	}
	userIDs := map[string]uuid.UUID{}
	for _, u := range users {
		existing, err := usersRepo.GetByEmail(ctx, u.Email)
		if err == nil {
			userIDs[u.Username] = existing.ID
			continue
		}
		if !repository.IsNotFound(err) {
			return err
		}
		hash, err := auth.HashPassword(u.Pwd)
		if err != nil {
			return err
		}
		usr := &model.User{Username: u.Username, Email: u.Email, PasswordHash: hash, Role: u.Role}
		if err := usersRepo.Create(ctx, usr); err != nil {
			return fmt.Errorf("create user %s: %w", u.Email, err)
		}
		userIDs[u.Username] = usr.ID
	}

	// ----- Projects -----
	projects := []model.Project{
		{Name: "payments-api", Description: "Backend pagamenti", RepositoryURL: "https://github.com/alloy/payments-api", WebhookToken: "tok_payments"},
		{Name: "web-dashboard", Description: "Dashboard clienti", RepositoryURL: "https://github.com/alloy/web-dashboard", WebhookToken: "tok_dashboard"},
		{Name: "notifications", Description: "Servizio notifiche", RepositoryURL: "https://github.com/alloy/notifications", WebhookToken: "tok_notifications"},
	}
	projectIDs := map[string]uuid.UUID{}
	for i := range projects {
		p := projects[i]
		existing, err := projectsRepo.GetByRepositoryURL(ctx, p.RepositoryURL)
		if err == nil {
			projectIDs[p.Name] = existing.ID
			continue
		}
		if !repository.IsNotFound(err) {
			return err
		}
		if err := projectsRepo.Create(ctx, &p); err != nil {
			return fmt.Errorf("create project %s: %w", p.Name, err)
		}
		projectIDs[p.Name] = p.ID
	}

	// ----- Releases (10 con stati diversi) -----
	now := time.Now()
	aliceID := userIDs["alice"]
	bobID := userIDs["bob"]

	type relSpec struct {
		Project string
		Branch  string
		Title   string
		Status  model.ReleaseStatus
		Who     uuid.UUID
		Age     time.Duration
	}
	releaseSpecs := []relSpec{
		{"payments-api", "feature/new-checkout", "Nuovo flow di checkout", model.ReleaseInProd, aliceID, 30 * 24 * time.Hour},
		{"payments-api", "feature/refund-api", "API rimborsi", model.ReleaseInCert, bobID, 5 * 24 * time.Hour},
		{"payments-api", "fix/idempotency", "Fix chiavi di idempotenza", model.ReleaseApproved, aliceID, 12 * time.Hour},
		{"payments-api", "feature/multi-currency", "Supporto multi-valuta", model.ReleaseDraft, bobID, time.Hour},
		{"web-dashboard", "feature/dark-mode", "Dark mode", model.ReleaseInDev, aliceID, 2 * 24 * time.Hour},
		{"web-dashboard", "feature/charts-v2", "Nuovi grafici", model.ReleaseInCert, bobID, 3 * 24 * time.Hour},
		{"web-dashboard", "fix/login-redirect", "Fix redirect post-login", model.ReleaseInProd, aliceID, 20 * 24 * time.Hour},
		{"notifications", "feature/sms-channel", "Canale SMS", model.ReleaseRejected, bobID, 7 * 24 * time.Hour},
		{"notifications", "feature/email-templates", "Template email", model.ReleaseInCert, aliceID, 24 * time.Hour},
		{"notifications", "feature/webhooks", "Outbound webhooks", model.ReleaseInDev, bobID, 6 * time.Hour},
	}

	for _, rs := range releaseSpecs {
		who := rs.Who
		rel := &model.Release{
			ProjectID:   projectIDs[rs.Project],
			BranchName:  rs.Branch,
			Title:       rs.Title,
			Description: "Demo release",
			Status:      rs.Status,
			CreatedBy:   &who,
		}
		if err := releasesRepo.Create(ctx, rel); err != nil {
			// se la release esiste già (run multiple), skip
			continue
		}

		// Deployment events coerenti con lo stato
		stages := stagesFor(rs.Status)
		for i, env := range stages {
			sha := fakeSHA(rs.Branch, i)
			ev := &model.DeploymentEvent{
				ReleaseID: rel.ID, Environment: env, CommitSHA: sha,
				DeployedBy: &who, Notes: "seeded",
			}
			if err := deploymentsRepo.Create(ctx, ev); err != nil {
				return err
			}
			if env == model.EnvCert {
				snaps := fakeCommits(rs.Branch, 3, now.Add(-rs.Age))
				if err := commitsRepo.BulkUpsert(ctx, rel.ID, snaps); err != nil {
					return err
				}
			}
		}

		// PR + check per release in_cert / approved / in_prod
		if rs.Status == model.ReleaseInCert || rs.Status == model.ReleaseApproved || rs.Status == model.ReleaseInProd {
			certSHA := fakeSHA(rs.Branch, indexOf(stages, model.EnvCert))
			pr := &model.PullRequest{
				ReleaseID: rel.ID,
				PRURL:     fmt.Sprintf("https://github.com/alloy/%s/pull/%d", rs.Project, 100+len(rs.Branch)),
				PRNumber:  100 + len(rs.Branch),
				// alcune PR hanno head divergente per simulare un check fallito
				HeadCommitSHA: chooseHead(rs.Branch, certSHA),
				BaseBranch:    "main",
				Status:        model.PROpen,
				OpenedAt:      now.Add(-rs.Age / 2),
			}
			if err := prsRepo.Create(ctx, pr); err != nil {
				return err
			}

			passed := pr.HeadCommitSHA == certSHA
			details := "head matches cert HEAD"
			if !passed {
				details = "head not present in cert snapshot"
			}
			ck := &model.CertificationCheck{
				PullRequestID: pr.ID, HeadCommitSHA: pr.HeadCommitSHA,
				CertCommitSHA: certSHA, Passed: passed, Details: details,
			}
			if err := checksRepo.Create(ctx, ck); err != nil {
				return err
			}
			if !passed {
				_ = prsRepo.UpdateStatus(ctx, pr.ID, model.PRBlocked)
			}
		}
	}

	_ = cfg
	return nil
}

func stagesFor(s model.ReleaseStatus) []model.Environment {
	switch s {
	case model.ReleaseDraft:
		return nil
	case model.ReleaseInDev:
		return []model.Environment{model.EnvDev}
	case model.ReleaseInCert, model.ReleaseApproved, model.ReleaseRejected:
		return []model.Environment{model.EnvDev, model.EnvCert}
	case model.ReleaseInProd:
		return []model.Environment{model.EnvDev, model.EnvCert, model.EnvProd}
	}
	return nil
}

func fakeSHA(branch string, n int) string {
	base := fmt.Sprintf("%x", []byte(branch))
	if len(base) < 30 {
		base = base + "0000000000000000000000000000000000000000"
	}
	return base[:38] + fmt.Sprintf("%02d", n%100)
}

func chooseHead(branch, certSHA string) string {
	// se branch contiene 'fix' simuliamo un commit aggiunto dopo cert → check fails
	for _, ch := range branch {
		if ch == 'f' {
			return certSHA[:30] + "deadbe" + "ef"
		}
	}
	return certSHA
}

func indexOf(envs []model.Environment, target model.Environment) int {
	for i, e := range envs {
		if e == target {
			return i
		}
	}
	return 0
}

func fakeCommits(branch string, n int, base time.Time) []model.CommitSnapshot {
	out := make([]model.CommitSnapshot, n)
	for i := 0; i < n; i++ {
		out[i] = model.CommitSnapshot{
			CommitSHA:     fakeSHA(branch, i),
			CommitMessage: fmt.Sprintf("[%s] commit %d", branch, i),
			Author:        "seed@diffinder.local",
			CommittedAt:   base.Add(-time.Duration(i) * time.Hour),
		}
	}
	return out
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
