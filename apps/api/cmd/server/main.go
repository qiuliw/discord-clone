package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/qiuliw/discord-clone/apps/api/internal/handler"
	"github.com/qiuliw/discord-clone/apps/api/internal/repository"
	"github.com/qiuliw/discord-clone/apps/api/internal/service"
)

func main() {
	addr := envOr("ADDR", ":8080")
	dbPath := envOr("DATABASE_PATH", "data/app.db")
	secretPath := envOr("SESSION_SECRET_PATH", "data/.session_secret")

	database, err := repository.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	secret, err := service.LoadOrCreateSecret(secretPath)
	if err != nil {
		log.Fatalf("session secret: %v", err)
	}

	auth := service.NewAuth(repository.NewUserRepository(database), secret)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.New(auth),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
