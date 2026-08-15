# Team Chat

Go monorepo：每个业务能力是一个独立服务，不设网关层。

```text
services/    → 微服务（业务边界）
pkg/         → 跨服务共享基础设施
proto/       → 服务间契约
deploy/      → 部署配置
scripts/     → 工具脚本
apps/        → 终端应用（web 等）
```

```text
.
├── apps/
│   └── web/
├── services/
│   └── user/
│       ├── cmd/server/          # 入口 / 依赖组装
│       ├── internal/
│       │   ├── handler/         # 请求处理
│       │   ├── service/         # 业务编排
│       │   ├── domain/          # 业务对象
│       │   └── repository/      # DB / sqlc
│       ├── migrations/
│       ├── sqlc.yaml
│       └── Dockerfile
├── pkg/
│   └── middleware/
├── proto/
├── deploy/
├── scripts/
├── docker-compose.yml
├── go.mod
└── go.sum
```

**边界：** Service 是业务能力；`cmd` 是运行入口；`internal` 是服务私有实现；`pkg` 只放通用基础能力；服务之间通过 HTTP / gRPC / 消息通信，不直接依赖彼此的 `internal`。

## Run

Needs Go 1.25+ and pnpm.

```bash
make all          # user :8080 + UI :3000
make user
make web
```

Or `docker compose up --build`.

Open [http://localhost:3000](http://localhost:3000). Accounts are stored in `data/app.db`.
