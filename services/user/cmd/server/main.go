package main

import (
	"log"
	"net/http"
	"time"

	"github.com/qiuliw/discord-clone/pkg/middleware"
	"github.com/qiuliw/discord-clone/services/user/internal/config"
	"github.com/qiuliw/discord-clone/services/user/internal/handler"
	"github.com/qiuliw/discord-clone/services/user/internal/repository"
	"github.com/qiuliw/discord-clone/services/user/internal/service"
)

func main() {
	cfg := config.Load()

	database, err := repository.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	secret, err := service.LoadOrCreateSecret(cfg.JWTSecretPath)
	if err != nil {
		log.Fatalf("jwt secret: %v", err)
	}

	auth := service.NewAuth(repository.NewUserRepository(database), secret)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           middleware.CORS(cfg.CORSAllowedHosts, handler.New(handler.NewAuthHandler(auth))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("user service listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
