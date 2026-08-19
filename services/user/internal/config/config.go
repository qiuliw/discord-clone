package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr             string
	DatabasePath     string
	JWTSecretPath    string
	CORSAllowedHosts []string
}

func Load() Config {
	return Config{
		Addr:             envOr("ADDR", ":8080"),
		DatabasePath:     envOr("DATABASE_PATH", "data/app.db"),
		JWTSecretPath:    envOr("JWT_SECRET_PATH", "data/.jwt_secret"),
		CORSAllowedHosts: envHostsOr("CORS_ALLOWED_ORIGINS", "localhost", "127.0.0.1", "[::1]"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envHostsOr(key string, defaults ...string) []string {
	if v := os.Getenv(key); v != "" {
		var hosts []string
		for _, h := range strings.Split(v, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				hosts = append(hosts, h)
			}
		}
		if len(hosts) > 0 {
			return hosts
		}
	}
	return defaults
}
