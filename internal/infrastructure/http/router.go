package http

import (
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/config"
	"avito-test-pr-service/internal/infrastructure/http/handlers"
	"avito-test-pr-service/internal/infrastructure/http/middleware"
	"avito-test-pr-service/internal/infrastructure/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router *chi.Mux
	log    *logger.Logger

	prService   input.PRInputPort
	teamService input.TeamInputPort
	userService input.UserInputPort
}

func NewRouter(log *logger.Logger, prSvc input.PRInputPort, teamSvc input.TeamInputPort, userSvc input.UserInputPort) *Router {
	return &Router{
		router:      chi.NewRouter(),
		log:         log,
		prService:   prSvc,
		teamService: teamSvc,
		userService: userSvc,
	}
}

func (r *Router) Setup(cfg *config.Config) {
	r.router.Use(chiMiddleware.RequestID)
	r.router.Use(chiMiddleware.RealIP)
	r.router.Use(chiMiddleware.Recoverer)
	r.router.Use(middlewares.RequestLoggerMiddleware(r.log))
	r.router.Use(chiMiddleware.Timeout(10 * time.Second))

	r.router.Mount("/users", r.setupUserRoutes())
}

func (r *Router) setupUserRoutes() http.Handler {
	h := handlers.NewUserHandler(r.userService, r.prService, r.log)
	sub := chi.NewRouter()
	sub.Post("/setIsActive", h.SetIsActive)
	sub.Get("/getReview", h.GetReviews)
	return sub
}

func (r *Router) GetRouter() *chi.Mux { return r.router }
