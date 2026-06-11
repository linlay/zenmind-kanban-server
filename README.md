# zenmind-kanban-server

## 1. 项目简介

`zenmind-kanban-server` 是 ZenMind Kanban 的 Go 1.26 服务端，负责保存任务看板状态，并通过 WebSocket 同步 desktop 和 website。

## 2. 快速开始

### 前置要求

- Go 1.26
- Docker / Docker Compose 可选

### 本地启动

```bash
cp .env.example .env
set -a
source .env
set +a
go run ./cmd/server
```

`.env.example` 默认监听 `:8080`。本地调试时可按需改成其他端口。

### 测试

```bash
go test ./...
```

### 生成 SQLite 数据库

基础库包含完整表结构、索引、工作流目录、默认项目和默认看板：

```bash
scripts/create-sqlite-db.sh --db ./data/kanban.db
```

追加演示项目和 issues：

```bash
scripts/create-sqlite-db.sh --db ./data/kanban.db --demo --force
```

脚本默认拒绝覆盖已有数据库；需要重建时传入 `--force`，会同时清理 `.db-wal` 和 `.db-shm`。

## 3. 配置说明

本项目的主要配置放在 `.env`，配置契约放在 `.env.example`。`.env` 只用于本地真实值，不提交。

常用配置：

```text
ZENMIND_KANBAN_ADDR=:8080
ZENMIND_KANBAN_DB=./data/kanban.db
ZENMIND_KANBAN_ALLOWED_ORIGINS=https://your-domain.example
ZENMIND_KANBAN_TOKEN=
ZENMIND_KANBAN_STATIC_DIR=
ZENMIND_KANBAN_CONTAINER_DB=/data/kanban.db
ZENMIND_KANBAN_DOCKER_NETWORK=zenmind-kanban-net
```

域名未补充前，`ZENMIND_KANBAN_ALLOWED_ORIGINS` 可临时设为 `*`。上线后应替换成真实 website origin。
`ZENMIND_KANBAN_STATIC_DIR` 非空时，服务会同源托管前端静态文件，并对 SPA 路由回退到 `index.html`。

当前版本不读取 `configs/*.yml`。配置优先级为：代码默认值 < 环境变量。

## 4. 部署

### 创建内部网络

双 compose 部署依赖同一个外部 Docker network。首次部署前创建：

```bash
docker network create zenmind-kanban-net
```

### 启动服务

```bash
cp .env.example .env
docker compose up --build -d
```

当前 compose 不发布宿主机端口，只在 Docker 网络内暴露：

```text
kanban-server:8080
```

SQLite 数据通过 `kanban_data` volume 持久化到容器内 `/data/kanban.db`。

### 外部入口

server 不直接面向公网或宿主机。外部 Caddy/Nginx/Traefik 应加入 `zenmind-kanban-net`，并把域名流量转发到 website 容器：

```text
kanban-website:80
```

website nginx 会把 `/ws` 和 `/api/` 反代到 `kanban-server:8080`。

## 5. 运维

### 健康检查

```bash
docker compose exec kanban-server wget -qO- http://127.0.0.1:8080/healthz
```

### Issue 查询性能检查

`GET /api/issues?projectId=<id>` 返回当前 project 及其子 project 下的完整 issue 列表，鉴权规则与 `/api/snapshot` 一致。响应头包含 `Server-Timing` 和 `X-Issue-Count`，可用于快速判断服务端查询耗时和返回数量。

```bash
curl -sS -o /dev/null \
  -w "code=%{http_code} time=%{time_total} size=%{size_download}\n" \
  -H "Authorization: Bearer $ZENMIND_KANBAN_TOKEN" \
  "https://kanban.zenmind.cc/api/issues?projectId=default"
```

### 查看日志

```bash
docker compose logs -f kanban-server
```

### 常见排查

- compose 启动失败：检查 `zenmind-kanban-net` 是否已创建。
- 外部无法访问 server：这是预期行为，server 不发布宿主端口。
- website 无法连接 server：确认 server 和 website compose 使用同一个 `ZENMIND_KANBAN_DOCKER_NETWORK`。
- WebSocket 被拒绝：检查 `ZENMIND_KANBAN_ALLOWED_ORIGINS` 是否包含真实域名 origin。
- 返回 unauthorized：检查 `ZENMIND_KANBAN_TOKEN` 是否与客户端传入 token 一致。
- desktop 离线：desktop 应连接 `wss://<你的域名>/ws?role=desktop`。
