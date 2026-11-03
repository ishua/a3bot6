# a3bot6 — routine operations via Telegram bots (multi-service, mcore-centered)
Map of all projects and agent rules: ../hq/AGENTS.md

## Purpose
A set of small services that perform routine operations triggered by Telegram commands.
mcore is the center: it receives messages, parses commands, stores all state and hands
out tasks; workers execute tasks and report back. Services are stateless and talk over
HTTP REST; state lives in mcore's SQLite database. The repository is public — watch for
leaking secrets. The sole developer understands English but prefers communication in Russian.

## Stack
- Go workspace (`go.work`): mcore, tbot, notes. Python 3: ytd2feed (yt-dlp).
- SQLite via mattn/go-sqlite3 — mcore builds/tests need CGO_ENABLED=1.
- Key libs: telegram-bot-api v5 (tbot), go-git (notes), cristalhq/aconfig + YAML configs,
  custom logger (`mcore/pkg/logger`: Debug/Info/Warn/Fatal, SetLogLevel).

## Structure
- `mcore/` — core: REST API (`internal/rest`), command routing (`internal/routing`),
  task processing (`internal/taskmng`), SQLite storage (`internal/storage/msqlclient`);
  shared types in `pkg/schema` (Task, Dialog, Message); worker client library in `pkg/mcoreclient`.
- `tbot/` — Telegram gateway: reads TG messages → mcore; receives jobs from mcore → sends to TG.
- `notes/` — notes worker: git-based notes (fsnotes), `/note` commands; task-list parsing
  in `internal/domain/taskparser`.
- `ytd2feed/` — Python worker: downloads YouTube audio and generates an RSS feed.
- `compose.yml`, `*/Dockerfile` — docker images; `.github/workflows/` — build and push to ghcr.io on `v*.*.*` tags.
- `conf/*.yaml` in each service — placeholder configs; real values exist only on the server.
- `docs/`, `static/` — work plans and bot assets.

## Task flow and REST API
- Flow: user writes in TG → tbot `POST /add-msg/` → mcore creates a dialog, parses the
  command, creates a task → worker `POST /get-task/` → executes → `POST /report-task/`.
- Endpoints: `POST /add-msg/`, `POST /get-task/`, `POST /report-task/`, `GET /health/`,
  `POST /delete-all-data/` (dev only). All except `/health/` require the `secret` header.
- Task types (`mcore/pkg/schema/task.go`): Msg, Ytdl, Rest, Note, Torrent, Finance, Syno;
  status flow Create → Sended → Done / Error.
- Workers talk to mcore only via `mcore/pkg/mcoreclient`: `NewClient(addr, secret)`,
  `AddMsg()`, `GetTask()`, `ReportTask()`, `ListeningTasks()` (background polling loop).

## How to run and verify
From the repo root:
- `make build` — `CGO_ENABLED=1 go build ./...` in each Go module (mcore, tbot, notes)
- `make test` — `go test ./...` in each Go module
- `make help` — target and module list.
ytd2feed is Python and is not covered by these targets. Per-module commands:
`cd mcore && go build ./...` (same for tbot, notes). Local run of the core:
`make -C mcore run` — builds `mcore/build/mcore` and starts it with `mcore/conf/mcore_config.yaml`.
ytd2feed: `pip install -r ytd2feed/requirements.txt && python ytd2feed/main.py`; it has no automated tests.

## Deploy
`docker-compose build` builds images locally; GitHub Actions pushes images to ghcr.io on
`v*.*.*` tags. The services actually run on host b; deploy goes through the bserv
repository (upd.sh per service folder). Any deploy — only with owner confirmation (hq/AGENTS.md).

## Code style
- Imports: stdlib first, then external; blank import `_` for side effects (sqlite3 driver).
- Naming: camelCase for variables/functions, PascalCase for exported, lowercase unexported;
  interfaces end in `-er` or a descriptive name; packages short and lowercase, no underscores.
- Errors: return explicitly, wrap with context `fmt.Errorf("action: %w", err)`;
  log at the call site; `logger.Fatal` only for fatal startup/config problems.
- Configs: aconfig with `default`/`required`/`usage` tags, YAML files in `conf/`, validated at startup.
- Python: stdlib imports first, snake_case functions/variables, PascalCase classes.

## Rules and restrictions
- No secrets in git (public repo): `conf/*.yaml` hold placeholders (`test`, `sometoken`) —
  keep them placeholders; real tokens live only on the server (.env / server configs).
- Workers must reach mcore through `mcoreclient` with the `secret` header — do not bypass.
- `pkg/schema` is the contract between mcore and workers: a task-type change touches
  routing, taskmng and the worker side — keep the flow above intact.

## Relations
- bserv — mcore, tbot, ytd (ytd2feed) and notes containers run on host b
  (hq/hosts.md: `b`, b.a3b.me); the RSS feed is served via caddy on b.
- fsnotes — the notes worker reads and syncs the owner's notes repository.
