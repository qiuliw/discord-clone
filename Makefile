.PHONY: user web all sqlc

sqlc:
	cd services/user && sqlc generate

user:
	go run ./services/user/cmd/server

web:
	cd apps/web && pnpm dev

all:
	$(MAKE) -j2 user web
