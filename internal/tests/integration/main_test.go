package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var (
	TestCtx context.Context
	PGC     *PostgresContainer
)

func TestMain(m *testing.M) {
	TestCtx = context.Background()
	var err error

	PGC, err = StartPostgres(TestCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres failed: %v\n", err)
		os.Exit(1)
	}

	if err := ApplyMigrations(TestCtx, PGC.DSN); err != nil {
		fmt.Fprintf(os.Stderr, "apply migrations failed: %v\n", err)
		_ = PGC.Close(TestCtx)
		os.Exit(1)
	}

	code := m.Run()
	_ = PGC.Close(TestCtx)
	os.Exit(code)
}
