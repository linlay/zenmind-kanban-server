# zenmind-kanban-server Agent Notes

## Architecture

`zenmind-kanban-server` is the canonical kanban state source for ZenMind. Desktop connects as a local capability client, while website connects as a remote UI client.

- `cmd/server/main.go`: HTTP server, health endpoint, snapshot endpoint, and WebSocket entrypoint.
- `internal/config`: environment-variable config loading.
- `internal/kanban`: issue model, validation, status machine, run-lock rules, and position calculations.
- `internal/store`: SQLite persistence, migrations, revisions, and event log writes.
- `internal/realtime`: WebSocket sessions, RPC envelopes, broadcasts, and desktop command forwarding.
- `internal/desktop`: desktop capability constants and protocol-facing declarations.

Container deployment uses an external Docker network. The service name `kanban-server` is the internal address, and the container listens on port `8080`.

## Interfaces

HTTP endpoints:

- `GET /healthz`: returns service health.
- `GET /api/snapshot`: returns the default board snapshot. Requires `Authorization: Bearer <token>` when `ZENMIND_KANBAN_TOKEN` is set.
- `GET /api/issues?projectId=<id>`: returns all issues under the selected project subtree. Requires `Authorization: Bearer <token>` when `ZENMIND_KANBAN_TOKEN` is set. Responses include `Server-Timing` and `X-Issue-Count` headers for quick performance checks.
- `GET /ws`: upgrades to WebSocket. Clients pass `role=web` or `role=desktop`; token may be passed by query param or bearer header.

In production, browsers and desktop clients should reach these endpoints through the website domain:

- `https://<domain>/api/snapshot`
- `https://<domain>/api/issues?projectId=default`
- `wss://<domain>/ws?role=web`
- `wss://<domain>/ws?role=desktop`

The website nginx proxies `/api/` and `/ws` to `http://kanban-server:8080` over the Docker network.

WebSocket envelope:

```json
{
  "v": 1,
  "type": "rpc.req",
  "id": "request-id",
  "op": "kanban.issue.create",
  "role": "web",
  "boardId": "default",
  "revision": 0,
  "payload": {}
}
```

Website operations:

- `kanban.snapshot.get`
- `kanban.issue.create`
- `kanban.issue.update`
- `kanban.issue.delete`
- `kanban.issue.move`
- `kanban.issue.assignAndRun`
- `kanban.automation.sync`
- `desktop.assistant.listAgents`

Desktop operations:

- `desktop.hello`
- `desktop.assistant.event`
- RPC responses for `desktop.assistant.startRun`, `desktop.assistant.listAgents`, and `desktop.automation.sync`

Server broadcasts:

- `kanban.snapshot`
- `kanban.issue.created`
- `kanban.issue.updated`
- `kanban.issue.deleted`
- `kanban.desktop.status`

## Data Rules

- v1 uses one board: `boardId=default`.
- Valid statuses: `backlog`, `todo`, `in_progress`, `completed`.
- Valid priorities: `high`, `medium`, `low`.
- Valid run states: `running`, `completed`, `failed`, `cancelled`.
- Issues with an active `runId` cannot be moved or directly status-switched.
- Normal update cannot directly mark an incomplete issue as `completed`.
- Completion is allowed through a move operation or a desktop assistant event.
- Drag sorting uses floating-point `position`.
- SQLite tables are `task_board_issues`, `task_board_events`, `task_board_meta`, and `desktop_clients`.

## Configuration Notes

- `.env.example` is the committed configuration contract.
- `.env` is for local values and must remain ignored by Git.
- Compose must not publish server ports to the host; use `expose: ["8080"]` only.
- Server and website compose files must use the same external Docker network, defaulting to `zenmind-kanban-net`.
- `ZENMIND_KANBAN_ALLOWED_ORIGINS` should be the real website origin once the domain is known; `*` is only a temporary pre-domain value.
- v1 does not read `configs/*.yml`; do not add empty `configs/` or `.gitkeep`.
- Current config precedence is code defaults < environment variables.
- If yml config is added later, implement explicit loading before documenting it: code defaults < yml < environment variables.
- Do not store real tokens, database files, or secrets in tracked files.
