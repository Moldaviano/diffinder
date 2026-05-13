package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/config"
	"github.com/alloy/diffinder/internal/handler"
	"github.com/alloy/diffinder/internal/logger"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/alloy/diffinder/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := repository.NewDB(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	deps := buildDeps(cfg, log, db)
	r := handler.NewRouter(deps)

	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, sCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer sCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}
	log.Info("server stopped")
}

func buildDeps(cfg *config.Config, log *slog.Logger, db *repository.DB) *handler.Deps {
	issuer := auth.NewIssuer(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	usersRepo := repository.NewUsersRepo(db)
	projectsRepo := repository.NewProjectsRepo(db)
	releasesRepo := repository.NewReleasesRepo(db)
	deploymentsRepo := repository.NewDeploymentsRepo(db)
	commitsRepo := repository.NewCommitsRepo(db)
	prsRepo := repository.NewPullRequestsRepo(db)
	checksRepo := repository.NewChecksRepo(db)
	dashRepo := repository.NewDashboardRepo(db)

	authSvc := service.NewAuthService(usersRepo, issuer)
	projectSvc := service.NewProjectService(projectsRepo)
	releaseSvc := service.NewReleaseService(releasesRepo, deploymentsRepo, commitsRepo, prsRepo)
	deploymentSvc := service.NewDeploymentService(releasesRepo, deploymentsRepo, commitsRepo)
	prSvc := service.NewPRService(prsRepo, checksRepo)
	checkSvc := service.NewCheckService(prsRepo, deploymentsRepo, commitsRepo, checksRepo)
	dashSvc := service.NewDashboardService(dashRepo, deploymentsRepo, prsRepo)
	webhookSvc := service.NewWebhookService(projectsRepo, releasesRepo, prsRepo, checkSvc)

	return &handler.Deps{
		Cfg: cfg, Logger: log, DB: db, Issuer: issuer,

		Users: usersRepo, Projects: projectsRepo, Releases: releasesRepo,
		Deployments: deploymentsRepo, Commits: commitsRepo, PRs: prsRepo,
		Checks: checksRepo, Dashboard: dashRepo,

		AuthSvc: authSvc, ProjectSvc: projectSvc, ReleaseSvc: releaseSvc,
		DeploymentSvc: deploymentSvc, PRSvc: prSvc, CheckSvc: checkSvc,
		DashSvc: dashSvc, WebhookSvc: webhookSvc,
	}
}
