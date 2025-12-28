// cmd/api/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/ogradyo/lotto-app/web"
    "github.com/ogradyo/lotto-app/internal/httpapi"
)

func main() {
	dsn := os.Getenv("LOTTO_DB_DSN")
	if dsn == "" {
		log.Fatal("LOTTO_DB_DSN is not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("error creating db pool: %v", err)
	}
	defer db.Close()

	tmpls := web.Templates
	//tmpls := template.Must(template.ParseFS(templateFS, "web/templates/*.html")) // key change

    server := httpapi.NewServer(db, tmpls)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
