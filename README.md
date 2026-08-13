# Team Chat

Go API in `apps/api`, Next.js in `apps/web`. Each app manages its own dependencies.

```
apps/
  api/     Go + SQLite
  web/     Next.js (pnpm)
```

## Run

Needs Go 1.22+ and pnpm.

```bash
# terminal 1 — API on :8080
make api

# terminal 2 — UI on :3000
cd apps/web && pnpm install && pnpm dev
```

Or `docker compose up --build`.

Open [http://localhost:3000](http://localhost:3000). Accounts are stored in `apps/api/data/app.db`.
