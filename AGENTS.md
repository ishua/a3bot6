# AGENTS.md - Development Guide for a3bot6

This file provides guidance for AI agents working on this codebase.

## Project Overview

Multi-service project:
The main goal is writing various services to perform routine operations via telegram
Services can be physically running on different servers

- GO
    - mcore - business logic core, message parsing, queue management and state storage
    - tbot - client for telegram, two threads: reads messages and pushes them to mcore, receives jobs for sending messages
    - synoc - client for managing the synology service
    - notes - client for writing notes in text format and syncing them via git
- python
    - ytd2feed - client for downloading audio from youtube and generating rss
    - tr_mng - client for managing transmission via api

Communication between services happens over http rest api
Data is stored in a SQLite database in mcore

The project sources are stored in a public repository, watch for leaking secrets and other private data.

The sole developer understands English but prefers communication in Russian

## Architecture

Service architecture with a shared SQLite DB and HTTP API. All services are stateless, state is stored in mcore.

### Components

| Service | Language | Purpose                              |
|----------|----------|--------------------------------------|
| mcore    | Go       | System center: REST API, storage, task management |
| tbot     | Go       | Telegram gateway: reads from TG → mcore, listens to mcore → sends to TG |
| ytd2feed | Python   | YouTube → RSS (yt-dlp)               |
| tr_mng   | Python   | Transmission API management          |
| synoc    | Go       | Synology Download Station API        |
| notes    | Go       | Git-based notes (fsnotes)            |

### Data Flow

1. User writes in Telegram
2. tbot → mcore (`POST /add-msg/`)
3. mcore: create dialog → parse the command → create task in the DB
4. Worker poll mcore (`POST /get-task/`) → execute → (`POST /report-task/`)

### Task System

- **TaskType**: Msg, Ytdl, Torrent, Note, Syno, Finance, Rest
- **TaskStatus flow**: Create → Sended → Done / Error
- Workers use `mcoreclient` library for communication

### REST API (mcore)

- `POST /add-msg/` - add a message from the user
- `POST /get-task/` - get a task for a worker (by type)
- `POST /report-task/` - update the task status
- `GET /health/` - health check
- `POST /delete-all-data/` - clean the DB (for dev)

All endpoints (except /health) require the `secret` header for authorization.

## Build, Test, and Run Commands

### Go Services (mcore, tbot, synoc, notes)

```bash
# Build mcore
cd mcore && go build -o build/mcore ./cmd/app

# Build tbot
cd tbot && go build -o build/tbot ./cmd/app

# Build synoc
cd synoc && go build -o build/synoc ./cmd

### Docker Builds

```bash
# Build all services
docker-compose build

# Build single service
docker build -t mcore -f ./mcore/Dockerfile .
docker build -t tbot -f ./tbot/Dockerfile .
docker build -t ytd -f ./ytd2feed/Dockerfile .

# Run all services
docker-compose up -d
```

## Code Style Guidelines

### Go (Primary Language)

#### Imports
- Use standard library imports first, then external packages
- Group stdlib (`fmt`, `os`, `time`, etc.) from external packages
- Use blank import (`_`) for side effects (e.g., `_ "github.com/mattn/go-sqlite3"`)

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/cristalhq/aconfig"
    "github.com/ishua/a3bot6/mcore/pkg/logger"
)
```

#### Naming Conventions
- **Variables/Functions**: camelCase (`newSqlClient`, `cfg`)
- **Exported Types/Constants**: PascalCase (`SqliteClient`, `MyConfig`)
- **Unexported**: lowercase (`dialogMng`, `taskMng`)
- **Interfaces**: suffix with `er` or descriptive name (`taskWorker`, `dialogMng`)
- **Packages**: short, lowercase, no underscores (`routing`, `taskmng`)

#### Types and Structs
- Use struct tags for serialization (JSON, YAML, DB)
- Prefer interfaces for dependency injection
- Use pointers for mutable receivers (`func (c *Client) DoTask(...)`)
- Define constants in a `const` block at package level

#### Error Handling
- Return errors explicitly, don't use exceptions
- Wrap errors with context: `fmt.Errorf("action description: %w", err)`
- Use sentinel errors for known conditions when appropriate
- Log errors at call site, not in utility functions

```go
// Good
if err != nil {
    return mr, fmt.Errorf("addmsg doPost: %w", err)
}

// Config validation with fatal exit
if len(cfg.Secrets) == 0 {
    logger.Fatal("no secrets configured")
}
```

#### Configuration
- Use `aconfig` library with YAML files
- Define config struct with tags: `default:"value"`, `required:"true"`, `usage:"description"`
- Config files stored in `conf/` directories per service

```go
type MyConfig struct {
    HttpPort string   `default:"8080" usage:"port for HTTP REST"`
    Debug    bool     `default:"false" usage:"turn on debug mode"`
}
```

### Python

#### Imports
- Standard library first, then third-party, then local
- Use absolute imports (`from app.config import Conf`)

#### Naming
- snake_case for functions, variables, properties
- PascalCase for classes

```python
class Conf():
    @property
    def mcore_addr(self) -> str:
        return self.conf.get("mcoreAddr", "http://localhost:8080")
```

## Project Structure

```
a3bot6/
├── mcore/                  # Core service - REST API, storage, task management
│   ├── cmd/app/            # Entry points
│   ├── internal/
│   │   ├── rest/           # HTTP handlers
│   │   ├── routing/        # Message routing and command parsing
│   │   ├── taskmng/        # Task creation and processing
│   │   ├── dialogmng/      # Dialog management
│   │   ├── storage/        # SQLite client and queries
│   │   └── functions/      # Utility functions
│   └── pkg/
│       ├── schema/         # Data structures (Task, Dialog, Message)
│       ├── logger/         # Custom logger
│       └── mcoreclient/    # HTTP client for mcore
├── tbot/                   # Telegram bot service
├── synoc/                  # Synology worker service
├── notes/                  # Notes service (Git-based)
├── ytd2feed/               # YouTube to RSS Python service
├── tr_mng/                 # Transmission management Python service
└── compose.yml             # Docker Compose configuration
```

## Important Patterns

### Dependency Injection
Interfaces are used for dependencies between packages (see `routing.go`).

### mcoreclient Library
All workers use `mcoreclient` for communication with mcore:
- `ListeningTasks()` - background process for polling tasks
- `AddMsg()` - sending messages to mcore
- `GetTask()` / `ReportTask()` - working with tasks

## Configuration Files

- `mcore/conf/mcore_config.yaml` - Core service config
- `tbot/conf/tbot_config.yaml` - Telegram bot config
- `ytd2feed/conf/ytdl_config.yaml` - YouTube downloader config
- `synoc/conf/config.yaml` - Synology worker config

## Logging

Use the custom logger package in mcore (`pkg/logger/logger.go`):
- `logger.Debug(msg)`
- `logger.Info(msg)`
- `logger.Warn(msg)`
- `logger.Fatal(msg)` - exits program
- `logger.SetLogLevel(logger.DEBUG)` - enable debug mode