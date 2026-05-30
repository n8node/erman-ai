package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/server"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	seedAdmin := flag.Bool("seed-admin", false, "create or update superadmin user")
	seedEmail := flag.String("email", "", "superadmin email address")
	seedPassword := flag.String("password", "", "superadmin password")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if cfg.Environment == "development" {
		if err := runMigrations(cfg.DatabaseURL); err != nil {
			logger.Warn("migrations skipped or failed", "error", err)
		}
	}

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if *seedAdmin {
		if *seedEmail == "" || *seedPassword == "" {
			logger.Error("seed-admin requires --email and --password")
			os.Exit(1)
		}
		authMW := middleware.NewAuth(cfg.JWTSecret)
		userRepo := repository.NewUserRepository(db.Pool)
		smtpRepo := repository.NewSMTPSettingsRepository(db.Pool)
		telegramRepo := repository.NewTelegramSettingsRepository(db.Pool)
		telegramSupportRepo := repository.NewTelegramSupportThreadRepository(db.Pool)
		telegramUserStateRepo := repository.NewTelegramUserStateRepository(db.Pool)
		telegramUrgentSendRepo := repository.NewTelegramUrgentSendRepository(db.Pool)
		maxSettingsRepo := repository.NewMaxSettingsRepository(db.Pool)
		tokenRepo := repository.NewEmailVerificationTokenRepository(db.Pool)
		passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(db.Pool)
		smtpSvc := service.NewSMTPSettingsService(smtpRepo)
		mailSvc := service.NewMailService(smtpSvc)
		telegramAssets := service.NewTelegramAssets(cfg.TelegramAssetsDir)
		telegramSettingsSvc := service.NewTelegramSettingsService(telegramRepo, telegramAssets)
		maxSettingsSvc := service.NewMaxSettingsService(maxSettingsRepo)
		maxSvc := service.NewMaxService(maxSettingsSvc, logger)
		telegramSvc := service.NewTelegramService(
			telegramSettingsSvc, telegramSupportRepo, telegramUserStateRepo, telegramUrgentSendRepo,
			mailSvc, maxSvc, telegramAssets, logger,
		)
		verifySvc := service.NewEmailVerificationService(userRepo, tokenRepo, mailSvc, telegramSvc, cfg)
		passwordResetSvc := service.NewPasswordResetService(userRepo, passwordResetTokenRepo, mailSvc, cfg)
		authSvc := service.NewAuthService(userRepo, authMW, verifySvc, passwordResetSvc, telegramSvc)
		user, err := authSvc.SeedAdmin(ctx, *seedEmail, *seedPassword)
		if err != nil {
			logger.Error("seed admin failed", "error", err)
			os.Exit(1)
		}
		logger.Info("superadmin ready", "email", user.Email, "id", user.ID)
		os.Exit(0)
	}

	srv := server.New(cfg, db, logger)
	httpServer := &http.Server{
		Addr:         srv.Addr(),
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 960 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", httpServer.Addr, "version", config.Version)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	srv.Shutdown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "error", err)
	}
	logger.Info("server stopped")
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "./migrations")
}
