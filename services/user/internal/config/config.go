package config

import "os"

type Config struct {
	Addr          string
	DatabasePath  string
	JWTSecretPath string
}

func Load() Config {
	return Config{
		Addr:          envOr("ADDR", ":8080"),
		DatabasePath:  envOr("DATABASE_PATH", "data/app.db"),
		JWTSecretPath: envOr("JWT_SECRET_PATH", "data/.jwt_secret"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
