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
	proposalReqRepo := repository.NewProposalRequestRepository(db.Pool)
	tooltipRepo := repository.NewTooltipRepository(db.Pool)
	translationRepo := repository.NewTranslationRepository(db.Pool)
	budgetConfigRepo := repository.NewCalculatorBudgetConfigRepository(db.Pool)
	strategyLLMRepo := repository.NewStrategyLLMSettingsRepository(db.Pool)
	smtpSettingsRepo := repository.NewSMTPSettingsRepository(db.Pool)
	telegramSettingsRepo := repository.NewTelegramSettingsRepository(db.Pool)
	emailTokenRepo := repository.NewEmailVerificationTokenRepository(db.Pool)

	usageLogRepo := repository.NewUsageLogRepository(db.Pool)

	smtpSettingsSvc := service.NewSMTPSettingsService(smtpSettingsRepo)
	mailSvc := service.NewMailService(smtpSettingsSvc)
	telegramSettingsSvc := service.NewTelegramSettingsService(telegramSettingsRepo)
	telegramSvc := service.NewTelegramService(telegramSettingsSvc)
	emailVerifySvc := service.NewEmailVerificationService(userRepo, emailTokenRepo, mailSvc, telegramSvc, cfg)
	authSvc := service.NewAuthService(userRepo, authMW, emailVerifySvc, telegramSvc)
	billingSvc := service.NewBillingService(planRepo, runRepo, userRepo)
	calcSvc := service.NewCalculatorService(cfg, runRepo, planRepo)
	shareSvc := service.NewShareService(sharedRepo, runRepo, billingSvc)
	leadSvc := service.NewLeadService(leadRepo, runRepo)
	proposalReqSvc := service.NewProposalRequestService(proposalReqRepo, runRepo, userRepo)
	runSvc := service.NewRunService(runRepo, planRepo)
	tooltipSvc := service.NewTooltipService(tooltipRepo)
	translationSvc := service.NewTranslationService(translationRepo)
	budgetConfigSvc := service.NewCalculatorBudgetConfigService(budgetConfigRepo)
	llmSvc := service.NewLLMService(cfg)
	strategyLLMSvc := service.NewStrategyLLMSettingsService(strategyLLMRepo, cfg, llmSvc)
	strategySvc := service.NewStrategyService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	proposalSvc := service.NewProposalService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	auditSvc := service.NewAuditService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	planSvc := service.NewPlanService(planRepo)
	adminUserSvc := service.NewAdminUserService(userRepo, planRepo, authMW, telegramSvc)

	authHandler := handler.NewAuthHandler(authSvc, authMW, cfg)
	calcHandler := handler.NewCalculatorHandler(calcSvc, billingSvc, authSvc, cfg)
	runsHandler := handler.NewRunsHandler(runSvc)
	tooltipHandler := handler.NewTooltipHandler(tooltipSvc, authSvc)
	translationHandler := handler.NewTranslationHandler(translationSvc)
	budgetConfigHandler := handler.NewCalculatorBudgetConfigHandler(budgetConfigSvc)
	strategyLLMHandler := handler.NewStrategyLLMSettingsHandler(strategyLLMSvc)
	strategyHandler := handler.NewStrategyHandler(strategySvc, authSvc, billingSvc)
	proposalHandler := handler.NewProposalHandler(proposalSvc, authSvc, billingSvc)
	auditHandler := handler.NewAuditHandler(auditSvc, authSvc)
	planHandler := handler.NewPlanHandler(planSvc)
	adminUserHandler := handler.NewAdminUserHandler(adminUserSvc, authMW, cfg)
	smtpHandler := handler.NewSMTPSettingsHandler(smtpSettingsSvc, mailSvc)
	telegramHandler := handler.NewTelegramSettingsHandler(telegramSettingsSvc, telegramSvc)
	shareHandler := handler.NewShareHandler(shareSvc, authSvc, billingSvc, cfg, runRepo)
	leadHandler := handler.NewLeadHandler(leadSvc, authSvc)
	proposalReqHandler := handler.NewProposalRequestHandler(proposalReqSvc)
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
		api.Get("/public/tooltips", tooltipHandler.ListAnonymous)
		api.Get("/public/translations", translationHandler.ListPublic)

		api.Route("/auth", func(auth chi.Router) {
			auth.Use(authRL.Middleware)
			auth.Post("/register", authHandler.Register)
			auth.Post("/verify-email", authHandler.VerifyEmail)
			auth.Post("/resend-verification", authHandler.ResendVerification)
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
			protected.Get("/tools/calculator/budget-config", budgetConfigHandler.GetPublic)
			protected.Post("/tools/calculator/run", calcHandler.Run)
			protected.Post("/tools/calculator/export", calcHandler.Export)
			protected.Post("/tools/calculator/proposal-request", proposalReqHandler.Create)
			protected.Post("/tools/strategy/run", strategyHandler.Run)
			protected.Post("/tools/strategy/export", strategyHandler.Export)
			protected.Post("/tools/proposal/run", proposalHandler.Run)
			protected.Post("/tools/proposal/export", proposalHandler.Export)
			protected.Post("/tools/audit/run", auditHandler.Run)
			protected.Get("/runs/{id}/stream", strategyHandler.Stream)

			protected.Get("/runs", runsHandler.List)
			protected.Get("/runs/{id}", runsHandler.Get)
			protected.Delete("/runs/{id}", runsHandler.Delete)
			protected.Post("/runs/{id}/share", shareHandler.Create)
			protected.Post("/leads", leadHandler.Create)

			protected.Get("/billing/plan", billingHandler.Plan)
			protected.Get("/billing/plans", billingHandler.ListPlans)
			protected.Post("/billing/switch", billingHandler.SwitchPlan)
			protected.Get("/tooltips", tooltipHandler.ListPublic)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(authMW.Required)
			admin.Use(authMW.SuperAdmin)
			admin.Get("/tooltips", tooltipHandler.ListAdmin)
			admin.Put("/tooltips", tooltipHandler.BulkUpdateAdmin)
			admin.Put("/tooltips/{key}", tooltipHandler.UpdateAdmin)
			admin.Get("/plans/meta", planHandler.Meta)
			admin.Get("/plans", planHandler.List)
			admin.Get("/plans/{id}", planHandler.Get)
			admin.Post("/plans", planHandler.Create)
			admin.Put("/plans/{id}", planHandler.Update)
			admin.Get("/users", adminUserHandler.List)
			admin.Patch("/users/{id}", adminUserHandler.UpdatePlan)
			admin.Delete("/users/{id}", adminUserHandler.Delete)
			admin.Post("/users/{id}/impersonate", adminUserHandler.Impersonate)
			admin.Get("/proposal-requests", proposalReqHandler.ListAdmin)
			admin.Get("/proposal-requests/{id}", proposalReqHandler.GetAdmin)
			admin.Patch("/proposal-requests/{id}", proposalReqHandler.UpdateStatus)
			admin.Get("/calculator-budget", budgetConfigHandler.GetAdmin)
			admin.Put("/calculator-budget", budgetConfigHandler.UpdateAdmin)
			admin.Get("/strategy-llm", strategyLLMHandler.GetAdmin)
			admin.Put("/strategy-llm", strategyLLMHandler.UpdateAdmin)
			admin.Post("/strategy-llm/test-connection", strategyLLMHandler.TestConnection)
			admin.Get("/translations", translationHandler.SearchAdmin)
			admin.Put("/translations", translationHandler.BulkUpsertAdmin)
			admin.Put("/translations/item", translationHandler.UpsertAdmin)
			admin.Delete("/translations/{key}", translationHandler.DeleteAdmin)
			admin.Get("/email-smtp", smtpHandler.GetAdmin)
			admin.Put("/email-smtp", smtpHandler.UpdateAdmin)
			admin.Post("/email-smtp/test", smtpHandler.SendTest)
			admin.Get("/telegram", telegramHandler.GetAdmin)
			admin.Put("/telegram", telegramHandler.UpdateAdmin)
			admin.Post("/telegram/test", telegramHandler.SendTest)
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
