package main

import (
	"avito-test-pr-service/internal/infrastructure/config"
	"avito-test-pr-service/internal/infrastructure/logger"
	"context"
	"fmt"
)

func main() {
	cfg := config.MustLoad()
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DbName)
	ctx := context.Background()
	log := logger.New(cfg.Env)

	_ = ctx
	_ = log
	_ = dsn

}
