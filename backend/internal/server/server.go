package server

import (
	"log/slog"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/handler"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg    *config.Config
	router chi.Router
}

func New(cfg *config.Config, db *repository.Postgres, logger *slog.Logger) *Server {
	r := chi.NewRouter()
	auth := middleware.NewAuth(cfg.JWTSecret)

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS("*"))

	health := handler.NewHealthHandler(cfg, db)
	r.Get("/health", health.ServeHTTP)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", health.ServeHTTP)
		// Phase 2+: auth, tools, runs, billing, admin
	})

	_ = auth // used in Phase 2

	return &Server{cfg: cfg, router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return ":" + s.cfg.ServerPort
}
