package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	httpapi "github.com/ogradyo/lotto-app/internal/http"
)

func main() {
	dsn := os.Getenv("LOTTO_DB_DSN")
	if dsn == "" {
		log.Fatal("LOTTO_DB_DSN is not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      httpapi.NewServer(db).Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
