package http

import (
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/config"
	"avito-test-pr-service/internal/infrastructure/logger"
	"context"
	"log/slog"
	"net/http"
)

type Server struct {
	address string
	log     *logger.Logger
	router  *Router
	server  *http.Server

	prService   input.PRInputPort
	teamService input.TeamInputPort
	userService input.UserInputPort
}

func NewServer(address string, log *logger.Logger, prSvc input.PRInputPort, teamSvc input.TeamInputPort, userSvc input.UserInputPort) *Server {
	return &Server{
		address:     address,
		log:         log,
		prService:   prSvc,
		teamService: teamSvc,
		userService: userSvc,
	}
}

func (s *Server) Run(cfg *config.Config) error {
	s.router = NewRouter(s.log, s.prService, s.teamService, s.userService)
	s.router.Setup(cfg)

	s.server = &http.Server{
		Addr:         s.address,
		Handler:      s.router.GetRouter(),
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	s.log.Info("Starting server", slog.String("address", s.address))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
