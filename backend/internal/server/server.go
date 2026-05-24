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
	authSvc := service.NewAuthService(userRepo, authMW)
	authHandler := handler.NewAuthHandler(authSvc, authMW, cfg)

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS("*"))

	health := handler.NewHealthHandler(cfg, db)
	r.Get("/health", health.ServeHTTP)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", health.ServeHTTP)

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
			})
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
