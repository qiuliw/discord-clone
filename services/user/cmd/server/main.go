package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/qiuliw/discord-clone/services/user/internal/handler"
	"github.com/qiuliw/discord-clone/services/user/internal/repository"
	"github.com/qiuliw/discord-clone/services/user/internal/service"
)

func main() {
	addr := envOr("ADDR", ":8080")
	dbPath := envOr("DATABASE_PATH", "data/app.db")
	secretPath := envOr("JWT_SECRET_PATH", "data/.jwt_secret")

	database, err := repository.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	secret, err := service.LoadOrCreateSecret(secretPath)
	if err != nil {
		log.Fatalf("jwt secret: %v", err)
	}

	auth := service.NewAuth(repository.NewUserRepository(database), secret)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.New(handler.NewAuthHandler(auth)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("user service listening on %s", addr)
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
