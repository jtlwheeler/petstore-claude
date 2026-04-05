package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jtlwheeler/petstore/internal/db"
	"github.com/jtlwheeler/petstore/internal/handlers"
	"github.com/jtlwheeler/petstore/internal/repository"
)

func main() {
	ctx := context.Background()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/petstore"
	}

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, "./migrations"); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	petRepo := repository.NewPetRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	router := handlers.SetupRoutes(petRepo, orderRepo, userRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
