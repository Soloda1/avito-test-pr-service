package integration

import (
	"avito-test-pr-service/internal/infrastructure/logger"
	"avito-test-pr-service/internal/infrastructure/migrator"
	"context"
	"net/url"
	"path/filepath"
	"runtime"
)

func ApplyMigrations(ctx context.Context, dsn string) error {
	log := logger.New("test")
	if parsed, err := url.Parse(dsn); err == nil {
		q := parsed.Query()
		if q.Get("sslmode") == "" {
			q.Set("sslmode", "disable")
			parsed.RawQuery = q.Encode()
			dsn = parsed.String()
		}
	}

	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(thisFile)
	migrationsPath := filepath.Clean(filepath.Join(baseDir, "../../..", "migrations"))
	m, err := migrator.NewMigrator(migrationsPath, dsn, log)
	if err != nil {
		return err
	}
	defer func() { _ = m.Close() }()
	return m.Up()
}
