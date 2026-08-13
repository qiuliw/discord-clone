.PHONY: api web

api:
	cd apps/api && go run ./cmd/server

web:
	cd apps/web && pnpm dev
