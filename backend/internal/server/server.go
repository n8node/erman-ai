package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/handler"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg      *config.Config
	router   chi.Router
	telegram *service.TelegramService
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
	projectInquiryRepo := repository.NewProjectInquiryRepository(db.Pool)
	inquiryTokenRepo := repository.NewInquiryVerificationTokenRepository(db.Pool)
	tooltipRepo := repository.NewTooltipRepository(db.Pool)
	translationRepo := repository.NewTranslationRepository(db.Pool)
	budgetConfigRepo := repository.NewCalculatorBudgetConfigRepository(db.Pool)
	strategyLLMRepo := repository.NewStrategyLLMSettingsRepository(db.Pool)
	legalRiskRepo := repository.NewLegalRiskRepository(db.Pool)
	legalScanLLMRepo := repository.NewLegalScanLLMSettingsRepository(db.Pool)
	smtpSettingsRepo := repository.NewSMTPSettingsRepository(db.Pool)
	paymentSettingsRepo := repository.NewPaymentSettingsRepository(db.Pool)
	telegramSettingsRepo := repository.NewTelegramSettingsRepository(db.Pool)
	telegramSupportThreadRepo := repository.NewTelegramSupportThreadRepository(db.Pool)
	telegramUserStateRepo := repository.NewTelegramUserStateRepository(db.Pool)
	telegramUrgentSendRepo := repository.NewTelegramUrgentSendRepository(db.Pool)
	maxSettingsRepo := repository.NewMaxSettingsRepository(db.Pool)
	yandexMetrikaSettingsRepo := repository.NewYandexMetrikaSettingsRepository(db.Pool)
	externalProjectRepo := repository.NewExternalProjectRepository(db.Pool)
	publicPageRepo := repository.NewPublicPageRepository(db.Pool)
	emailTokenRepo := repository.NewEmailVerificationTokenRepository(db.Pool)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(db.Pool)
	planCheckoutRepo := repository.NewPlanCheckoutRepository(db.Pool)

	usageLogRepo := repository.NewUsageLogRepository(db.Pool)

	smtpSettingsSvc := service.NewSMTPSettingsService(smtpSettingsRepo)
	paymentSettingsSvc := service.NewPaymentSettingsService(paymentSettingsRepo, cfg)
	mailSvc := service.NewMailService(smtpSettingsSvc)
	telegramAssets := service.NewTelegramAssets(cfg.TelegramAssetsDir)
	_ = telegramAssets.EnsureDir()
	telegramSettingsSvc := service.NewTelegramSettingsService(telegramSettingsRepo, telegramAssets)
	maxSettingsSvc := service.NewMaxSettingsService(maxSettingsRepo)
	yandexMetrikaSettingsSvc := service.NewYandexMetrikaSettingsService(yandexMetrikaSettingsRepo)
	maxSvc := service.NewMaxService(maxSettingsSvc, logger)
	maxSettingsSvc.BindRuntimeStatus(func() model.MaxBotRuntimeStatus {
		return maxSvc.CheckHealth(context.Background())
	})
	telegramSvc := service.NewTelegramService(
		telegramSettingsSvc,
		telegramSupportThreadRepo,
		telegramUserStateRepo,
		telegramUrgentSendRepo,
		mailSvc,
		maxSvc,
		telegramAssets,
		logger,
	)
	telegramSettingsSvc.BindRuntimeStatus(telegramSvc.GetRuntimeStatus)
	telegramSvc.Start()
	emailVerifySvc := service.NewEmailVerificationService(userRepo, emailTokenRepo, mailSvc, telegramSvc, cfg)
	passwordResetSvc := service.NewPasswordResetService(userRepo, passwordResetTokenRepo, mailSvc, cfg)
	authSvc := service.NewAuthService(userRepo, authMW, emailVerifySvc, passwordResetSvc, telegramSvc)
	billingSvc := service.NewBillingService(planRepo, runRepo, userRepo)
	checkoutSvc := service.NewCheckoutService(planCheckoutRepo, userRepo, planRepo, paymentSettingsSvc, telegramSvc, cfg)
	calcSvc := service.NewCalculatorService(cfg, runRepo, planRepo)
	shareSvc := service.NewShareService(sharedRepo, runRepo, billingSvc)
	leadSvc := service.NewLeadService(leadRepo, runRepo)
	proposalReqSvc := service.NewProposalRequestService(proposalReqRepo, runRepo, userRepo, telegramSvc, cfg)
	projectInquirySvc := service.NewProjectInquiryService(projectInquiryRepo, inquiryTokenRepo, runRepo, userRepo, shareSvc, mailSvc, telegramSvc, cfg)
	runSvc := service.NewRunService(runRepo, planRepo)
	tooltipSvc := service.NewTooltipService(tooltipRepo)
	translationSvc := service.NewTranslationService(translationRepo)
	budgetConfigSvc := service.NewCalculatorBudgetConfigService(budgetConfigRepo)
	llmSvc := service.NewLLMService(cfg)
	strategyLLMSvc := service.NewStrategyLLMSettingsService(strategyLLMRepo, cfg, llmSvc)
	legalScanLLMSvc := service.NewLegalScanLLMSettingsService(legalScanLLMRepo, strategyLLMSvc, cfg)
	legalRiskSvc := service.NewLegalRiskService(legalRiskRepo)
	strategySvc := service.NewStrategyService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	proposalSvc := service.NewProposalService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	auditSvc := service.NewAuditService(cfg, runRepo, planRepo, billingSvc, llmSvc, strategyLLMSvc, usageLogRepo, logger)
	legalScanSvc := service.NewLegalScanService(cfg, runRepo, planRepo, legalRiskRepo, billingSvc, llmSvc, legalScanLLMSvc, strategyLLMSvc, usageLogRepo, logger)
	planSvc := service.NewPlanService(planRepo)
	externalProjectSvc := service.NewExternalProjectService(externalProjectRepo)
	publicPageSvc := service.NewPublicPageService(publicPageRepo)
	adminUserSvc := service.NewAdminUserService(userRepo, planRepo, authMW, telegramSvc)

	authHandler := handler.NewAuthHandler(authSvc, authMW, cfg, projectInquiryRepo, proposalReqRepo)
	calcHandler := handler.NewCalculatorHandler(calcSvc, billingSvc, authSvc, cfg)
	runsHandler := handler.NewRunsHandler(runSvc)
	tooltipHandler := handler.NewTooltipHandler(tooltipSvc, authSvc)
	translationHandler := handler.NewTranslationHandler(translationSvc)
	budgetConfigHandler := handler.NewCalculatorBudgetConfigHandler(budgetConfigSvc)
	strategyLLMHandler := handler.NewStrategyLLMSettingsHandler(strategyLLMSvc)
	strategyHandler := handler.NewStrategyHandler(strategySvc, authSvc, billingSvc)
	proposalHandler := handler.NewProposalHandler(proposalSvc, authSvc, billingSvc)
	auditHandler := handler.NewAuditHandler(auditSvc, authSvc)
	legalScanHandler := handler.NewLegalScanHandler(legalScanSvc, authSvc, billingSvc)
	legalRiskAdminHandler := handler.NewLegalRiskAdminHandler(legalRiskSvc)
	legalScanLLMHandler := handler.NewLegalScanLLMSettingsHandler(legalScanLLMSvc)
	planHandler := handler.NewPlanHandler(planSvc)
	adminUserHandler := handler.NewAdminUserHandler(adminUserSvc, authMW, cfg)
	smtpHandler := handler.NewSMTPSettingsHandler(smtpSettingsSvc, mailSvc)
	paymentHandler := handler.NewPaymentSettingsHandler(paymentSettingsSvc)
	paymentWebhookHandler := handler.NewPaymentWebhookHandler(paymentSettingsSvc, checkoutSvc, logger)
	telegramHandler := handler.NewTelegramSettingsHandler(telegramSettingsSvc, telegramSvc)
	maxHandler := handler.NewMaxSettingsHandler(maxSettingsSvc, maxSvc)
	yandexMetrikaHandler := handler.NewYandexMetrikaSettingsHandler(yandexMetrikaSettingsSvc)
	externalProjectHandler := handler.NewExternalProjectHandler(externalProjectSvc)
	publicPageHandler := handler.NewPublicPageHandler(publicPageSvc)
	shareHandler := handler.NewShareHandler(shareSvc, authSvc, billingSvc, cfg, runRepo)
	leadHandler := handler.NewLeadHandler(leadSvc, authSvc)
	proposalReqHandler := handler.NewProposalRequestHandler(proposalReqSvc)
	projectInquiryHandler := handler.NewProjectInquiryHandler(projectInquirySvc, authSvc, cfg)
	inquiryRL := middleware.NewRateLimiter(5, time.Hour)
	billingHandler := handler.NewBillingHandler(billingSvc, checkoutSvc, planRepo, runRepo)
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
		api.Get("/public/calculator/budget-config", budgetConfigHandler.GetPublic)
		api.Get("/public/pages/slugs", publicPageHandler.ListSlugs)
		api.Get("/public/pages/{slug}", publicPageHandler.GetPublic)
		api.Get("/public/billing/plans", billingHandler.ListPlans)
		api.Get("/public/projects", externalProjectHandler.ListPublic)
		api.Get("/public/yandex-metrika", yandexMetrikaHandler.GetPublic)
		api.With(inquiryRL.Middleware).Post("/public/project-inquiries", projectInquiryHandler.CreatePublic)
		api.Get("/public/project-inquiries/verify", projectInquiryHandler.Verify)
		api.Post("/billing/yookassa/webhook", paymentWebhookHandler.YookassaWebhook)
		api.Get("/billing/robokassa/result", paymentWebhookHandler.RobokassaResult)
		api.Post("/billing/robokassa/result", paymentWebhookHandler.RobokassaResult)

		api.Route("/auth", func(auth chi.Router) {
			auth.Use(authRL.Middleware)
			auth.Post("/register", authHandler.Register)
			auth.Post("/verify-email", authHandler.VerifyEmail)
			auth.Post("/resend-verification", authHandler.ResendVerification)
			auth.Post("/forgot-password", authHandler.ForgotPassword)
			auth.Post("/reset-password", authHandler.ResetPassword)
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
			protected.Post("/tools/legal-scan/run", legalScanHandler.Run)
			protected.Post("/tools/legal-scan/export", legalScanHandler.Export)
			protected.Get("/runs/{id}/stream", strategyHandler.Stream)

			protected.Get("/runs", runsHandler.List)
			protected.Get("/runs/{id}", runsHandler.Get)
			protected.Delete("/runs/{id}", runsHandler.Delete)
			protected.Post("/runs/{id}/share", shareHandler.Create)
			protected.Post("/leads", leadHandler.Create)
			protected.Post("/project-inquiries", projectInquiryHandler.CreateAuthenticated)

			protected.Get("/billing/plan", billingHandler.Plan)
			protected.Get("/billing/plans", billingHandler.ListPlans)
			protected.Post("/billing/switch", billingHandler.SwitchPlan)
			protected.Post("/billing/checkout", billingHandler.CreateCheckout)
			protected.Get("/tooltips", tooltipHandler.ListPublic)
			protected.Get("/projects", externalProjectHandler.ListPublic)
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
			admin.Get("/project-inquiries", projectInquiryHandler.ListAdmin)
			admin.Get("/project-inquiries/{id}", projectInquiryHandler.GetAdmin)
			admin.Patch("/project-inquiries/{id}", projectInquiryHandler.UpdateStatus)
			admin.Get("/calculator-budget", budgetConfigHandler.GetAdmin)
			admin.Put("/calculator-budget", budgetConfigHandler.UpdateAdmin)
			admin.Get("/strategy-llm", strategyLLMHandler.GetAdmin)
			admin.Put("/strategy-llm", strategyLLMHandler.UpdateAdmin)
			admin.Post("/strategy-llm/test-connection", strategyLLMHandler.TestConnection)
			admin.Get("/legal-risks", legalRiskAdminHandler.List)
			admin.Post("/legal-risks", legalRiskAdminHandler.Create)
			admin.Put("/legal-risks/{risk_id}", legalRiskAdminHandler.Update)
			admin.Get("/legal-scan-llm", legalScanLLMHandler.GetAdmin)
			admin.Put("/legal-scan-llm", legalScanLLMHandler.UpdateAdmin)
			admin.Get("/translations", translationHandler.SearchAdmin)
			admin.Put("/translations", translationHandler.BulkUpsertAdmin)
			admin.Put("/translations/item", translationHandler.UpsertAdmin)
			admin.Delete("/translations/{key}", translationHandler.DeleteAdmin)
			admin.Get("/email-smtp", smtpHandler.GetAdmin)
			admin.Put("/email-smtp", smtpHandler.UpdateAdmin)
			admin.Post("/email-smtp/test", smtpHandler.SendTest)
			admin.Get("/payments", paymentHandler.GetAdmin)
			admin.Put("/payments", paymentHandler.UpdateAdmin)
			admin.Post("/payments/test", paymentHandler.TestConnection)
			admin.Get("/telegram", telegramHandler.GetAdmin)
			admin.Put("/telegram", telegramHandler.UpdateAdmin)
			admin.Get("/telegram/status", telegramHandler.GetStatus)
			admin.Post("/telegram/restart", telegramHandler.Restart)
			admin.Post("/telegram/test", telegramHandler.SendTest)
			admin.Post("/telegram/start-image", telegramHandler.UploadStartImage)
			admin.Get("/telegram/start-image", telegramHandler.GetStartImage)
			admin.Get("/max", maxHandler.GetAdmin)
			admin.Put("/max", maxHandler.UpdateAdmin)
			admin.Get("/max/status", maxHandler.GetStatus)
			admin.Post("/max/test", maxHandler.SendTest)
			admin.Get("/yandex-metrika", yandexMetrikaHandler.GetAdmin)
			admin.Put("/yandex-metrika", yandexMetrikaHandler.UpdateAdmin)
			admin.Get("/projects", externalProjectHandler.ListAdmin)
			admin.Post("/projects", externalProjectHandler.Create)
			admin.Put("/projects/{id}", externalProjectHandler.Update)
			admin.Delete("/projects/{id}", externalProjectHandler.Delete)
			admin.Get("/public-pages", publicPageHandler.ListAdmin)
			admin.Post("/public-pages", publicPageHandler.CreateAdmin)
			admin.Put("/public-pages/{id}", publicPageHandler.UpdateAdmin)
			admin.Delete("/public-pages/{id}", publicPageHandler.DeleteAdmin)
		})
	})

	return &Server{cfg: cfg, router: r, telegram: telegramSvc}
}

func (s *Server) Shutdown() {
	if s.telegram != nil {
		s.telegram.Stop()
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return ":" + s.cfg.ServerPort
}
