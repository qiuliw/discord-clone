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
# both services (API :8080, UI :3000)
make all

# or separately
make api
make web
```

Or `docker compose up --build`.

Open [http://localhost:3000](http://localhost:3000). Accounts are stored in `apps/api/data/app.db`.
