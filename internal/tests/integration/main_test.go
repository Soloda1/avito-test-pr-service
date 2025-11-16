package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var (
	testCtx context.Context
	pgC     *PostgresContainer
)

func TestMain(m *testing.M) {
	testCtx = context.Background()
	var err error

	pgC, err = StartPostgres(testCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres failed: %v\n", err)
		os.Exit(1)
	}

	if err := ApplyMigrations(testCtx, pgC.DSN); err != nil {
		fmt.Fprintf(os.Stderr, "apply migrations failed: %v\n", err)
		_ = pgC.Close(testCtx)
		os.Exit(1)
	}

	code := m.Run()
	_ = pgC.Close(testCtx)
	os.Exit(code)
}
