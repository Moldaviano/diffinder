package handler

import (
	"log/slog"
	"net/http"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/config"
	"github.com/alloy/diffinder/internal/httpx"
	appmw "github.com/alloy/diffinder/internal/middleware"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
	"github.com/alloy/diffinder/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Deps raccoglie l'intero grafo di dipendenze necessario al router.
// Costruito in cmd/server/main.go.
type Deps struct {
	Cfg    *config.Config
	Logger *slog.Logger
	DB     *repository.DB

	Issuer *auth.Issuer

	// Repos
	Users        *repository.UsersRepo
	Projects     *repository.ProjectsRepo
	Releases     *repository.ReleasesRepo
	Deployments  *repository.DeploymentsRepo
	Commits      *repository.CommitsRepo
	PRs          *repository.PullRequestsRepo
	Checks       *repository.ChecksRepo
	Dashboard    *repository.DashboardRepo

	// Services
	AuthSvc       *service.AuthService
	ProjectSvc    *service.ProjectService
	ReleaseSvc    *service.ReleaseService
	DeploymentSvc *service.DeploymentService
	PRSvc         *service.PRService
	CheckSvc      *service.CheckService
	DashSvc       *service.DashboardService
	WebhookSvc    *service.WebhookService
}

// NewRouter monta tutte le route REST + webhook.
func NewRouter(d *Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(appmw.Logger(d.Logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.Cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Hub-Signature-256"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		// Auth (public)
		r.Route("/auth", func(r chi.Router) {
			NewAuthHandler(d.AuthSvc).Mount(r)
		})

		// Webhooks (public, HMAC-protetto)
		r.Route("/webhooks", func(r chi.Router) {
			NewWebhookHandler(d.WebhookSvc, d.Cfg.Webhook.GitHubSecret).Mount(r)
		})

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(appmw.Auth(d.Issuer))

			r.Route("/projects", func(r chi.Router) {
				NewProjectsHandler(d.ProjectSvc, d.ReleaseSvc).Mount(r)
			})
			r.Route("/releases", func(r chi.Router) {
				NewReleasesHandler(d.ReleaseSvc, d.DeploymentSvc).Mount(r)
			})
			r.Route("/pull-requests", func(r chi.Router) {
				NewPullRequestsHandler(d.PRSvc, d.CheckSvc).Mount(r)
			})
			r.Route("/dashboard", func(r chi.Router) {
				NewDashboardHandler(d.DashSvc).Mount(r)
			})
			r.Route("/users", func(r chi.Router) {
				r.Use(appmw.RequireRole(model.RoleAdmin))
				NewUsersHandler(d.Users).Mount(r)
			})
		})
	})

	return r
}
