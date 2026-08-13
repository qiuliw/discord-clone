.PHONY: api web all

api:
	cd apps/api && go run ./cmd/server

web:
	cd apps/web && pnpm dev

all:
	$(MAKE) -j2 api web
