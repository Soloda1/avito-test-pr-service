package main

import (
	"avito-test-pr-service/internal/application/pr"
	teamapp "avito-test-pr-service/internal/application/team"
	userapp "avito-test-pr-service/internal/application/user"
	"avito-test-pr-service/internal/infrastructure/config"
	httpserver "avito-test-pr-service/internal/infrastructure/http"
	"avito-test-pr-service/internal/infrastructure/logger"
	pg_uow "avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/infrastructure/reviewerselector"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.MustLoad()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DbName,
	)

	ctx := context.Background()
	log := logger.New(cfg.Env)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("Failed to parse postgres pool config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("Failed to create postgres pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	uow := pg_uow.NewPostgresUOW(pool, log)
	selector := reviewerselector.NewRandomReviewerSelector()

	userService := userapp.NewService(uow, log)
	teamService := teamapp.NewService(uow, log)
	prService := pr.NewService(uow, selector, log)

	addr := fmt.Sprintf("%s:%d", cfg.HTTPServer.Address, cfg.HTTPServer.Port)
	server := httpserver.NewServer(addr, log, prService, teamService, userService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		if err := server.Run(cfg); err != nil {
			log.Error("HTTP server error", slog.String("error", err.Error()))
		}
		done <- true
	}()

	<-quit
	log.Info("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	<-done
	log.Info("Server exited")
}
