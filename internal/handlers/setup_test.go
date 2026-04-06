package handlers_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jtlwheeler/petstore/internal/db"
	"github.com/jtlwheeler/petstore/internal/handlers"
	"github.com/jtlwheeler/petstore/internal/repository"
	"github.com/jtlwheeler/petstore/internal/db/migrations"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("petstore_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic("starting postgres container: " + err.Error())
	}

	defer func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pgContainer.Terminate(timeoutCtx) //nolint:errcheck
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("getting connection string: " + err.Error())
	}

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		panic("connecting to database: " + err.Error())
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, migrations.FS); err != nil {
		panic("running migrations: " + err.Error())
	}

	petRepo := repository.NewPetRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	h := handlers.SetupRoutes(pool, petRepo, orderRepo, userRepo)
	testServer = httptest.NewServer(h)
	defer testServer.Close()

	os.Exit(m.Run())
}
