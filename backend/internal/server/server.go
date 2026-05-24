package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/handler"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg    *config.Config
	router chi.Router
}

func New(cfg *config.Config, db *repository.Postgres, logger *slog.Logger) *Server {
	r := chi.NewRouter()
	authMW := middleware.NewAuth(cfg.JWTSecret)
	authRL := middleware.NewRateLimiter(10, time.Minute)

	userRepo := repository.NewUserRepository(db.Pool)
	runRepo := repository.NewToolRunRepository(db.Pool)
	planRepo := repository.NewPlanRepository(db.Pool)
	sharedRepo := repository.NewSharedReportRepository(db.Pool)
	leadRepo := repository.NewLeadRepository(db.Pool)
	tooltipRepo := repository.NewTooltipRepository(db.Pool)

	authSvc := service.NewAuthService(userRepo, authMW)
	billingSvc := service.NewBillingService(planRepo, runRepo)
	calcSvc := service.NewCalculatorService(cfg, runRepo, planRepo)
	shareSvc := service.NewShareService(sharedRepo, runRepo, billingSvc)
	leadSvc := service.NewLeadService(leadRepo, runRepo)
	runSvc := service.NewRunService(runRepo, planRepo)
	tooltipSvc := service.NewTooltipService(tooltipRepo)

	authHandler := handler.NewAuthHandler(authSvc, authMW, cfg)
	calcHandler := handler.NewCalculatorHandler(calcSvc, billingSvc, authSvc, cfg)
	runsHandler := handler.NewRunsHandler(runSvc)
	tooltipHandler := handler.NewTooltipHandler(tooltipSvc, authSvc)
	shareHandler := handler.NewShareHandler(shareSvc, authSvc, billingSvc, cfg, runRepo)
	leadHandler := handler.NewLeadHandler(leadSvc, authSvc)
	billingHandler := handler.NewBillingHandler(billingSvc, planRepo, runRepo)
	toolsHandler := handler.NewToolsHandler(planRepo, runRepo, billingSvc)

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS("*"))

	health := handler.NewHealthHandler(cfg, db)
	r.Get("/health", health.ServeHTTP)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", health.ServeHTTP)
		api.Get("/shared/{token}", shareHandler.GetPublic)

		api.Route("/auth", func(auth chi.Router) {
			auth.Use(authRL.Middleware)
			auth.Post("/register", authHandler.Register)
			auth.Post("/login", authHandler.Login)
			auth.Post("/logout", authHandler.Logout)

			auth.Group(func(protected chi.Router) {
				protected.Use(authMW.Required)
				protected.Get("/me", authHandler.Me)
				protected.Put("/me", authHandler.UpdateMe)
				protected.Post("/change-password", authHandler.ChangePassword)
				protected.Post("/onboarding", authHandler.Onboarding)
			})
		})

		api.Group(func(protected chi.Router) {
			protected.Use(authMW.Required)

			protected.Get("/tools", toolsHandler.List)
			protected.Post("/tools/calculator/run", calcHandler.Run)
			protected.Post("/tools/calculator/export", calcHandler.Export)

			protected.Get("/runs", runsHandler.List)
			protected.Get("/runs/{id}", runsHandler.Get)
			protected.Delete("/runs/{id}", runsHandler.Delete)
			protected.Post("/runs/{id}/share", shareHandler.Create)
			protected.Post("/leads", leadHandler.Create)

			protected.Get("/billing/plan", billingHandler.Plan)
			protected.Get("/tooltips", tooltipHandler.ListPublic)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(authMW.Required)
			admin.Use(authMW.SuperAdmin)
			admin.Get("/tooltips", tooltipHandler.ListAdmin)
			admin.Put("/tooltips", tooltipHandler.BulkUpdateAdmin)
			admin.Put("/tooltips/{key}", tooltipHandler.UpdateAdmin)
		})
	})

	return &Server{cfg: cfg, router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return ":" + s.cfg.ServerPort
}
