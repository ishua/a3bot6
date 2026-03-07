# AGENTS.md - Development Guide for a3bot6

This file provides guidance for AI agents working on this codebase.

## Project Overview

Мультисервисный проект:
Основная цель, написание разных сервисов для выполнения рутинных операций через telegram
Сервисы могут быть физически запущенны на разных сервера

- GO
    - mcore - ядро бизнес логики, парсинг сообщений, управление очередями и хранение стейта
    - tbot - клиент для telegram, два потока: читает сообщения кидает в mcore, получает задания на отправку сообщений
    - synoc - клиент для управления synology сервисом
    - notes - клиент для написания заметок в текстовом формате и синком их через git
- python
    - ytd2feed - клиент для загрузки аудио из youtube и формирования rss
    - tr_mng - клиент для управления transmission через api

Коммуникация между сервисами происходит по http rest api
Данные хранятся в SQLite database в mcore

Исходники проекта хранятся в публичном репозитории, следить за утечкой секретов и других приватных данных.

Единственный разработчик понимает английский, но предпочитает общение на русском языке

## Architecture

Сервисная архитектура с общей SQLite БД и HTTP API. Все сервисы stateless, состояние хранится в mcore.

### Components

| Service  | Language | Purpose                              |
|----------|----------|--------------------------------------|
| mcore    | Go       | Центр системы: REST API, storage, task management |
| tbot     | Go       | Telegram шлюз: читает из TG → mcore, слушает mcore → отправляет в TG |
| ytd2feed | Python   | YouTube → RSS (yt-dlp)               |
| tr_mng   | Python   | Transmission API управление          |
| synoc    | Go       | Synology Download Station API        |
| notes    | Go       | Git-based заметки (fsnotes)          |

### Data Flow

1. Пользователь пишет в Telegram
2. tbot → mcore (`POST /add-msg/`)
3. mcore: создать dialog → распарсить команду → создать task в БД
4. Worker poll mcore (`POST /get-task/`) → выполнить → (`POST /report-task/`)

### Task System

- **TaskType**: Msg, Ytdl, Torrent, Note, Syno, Finance, Rest
- **TaskStatus flow**: Create → Sended → Done / Error
- Workers используют `mcoreclient` library для коммуникации

### REST API (mcore)

- `POST /add-msg/` - добавить сообщение от пользователя
- `POST /get-task/` - получить задачу для worker (по типу)
- `POST /report-task/` - обновить статус задачи
- `GET /health/` - health check
- `POST /delete-all-data/` - очистка БД (для dev)

Все endpoints (кроме /health) требуют header `secret` для авторизации.

## Build, Test, and Run Commands

### Go Services (mcore, tbot, synoc, notes)

```bash
# Build mcore
cd mcore && go build -o build/mcore ./cmd/app

# Build tbot
cd tbot && go build -o build/tbot ./cmd/app

# Build synoc
cd synoc && go build -o build/synoc ./cmd

# Run mcore (from mcore directory)
cd mcore && ./build/mcore

# Or use Makefile in mcore
cd mcore && make build
cd mcore && make run
```

### Single Test Execution (Go)

```bash
# Run a single test
go test -v -run TestFunctionName ./...

# Run tests in specific package
go test -v -run TestFunctionName ./internal/storage/...
```

### Python Services (ytd2feed, tr_mng)

```bash
# Install dependencies
pip install -r ytd2feed/requirements.txt

# Run ytd2feed
cd ytd2feed && python main.py

# Run tr_mng
cd tr_mng && python main.py
```

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
Interfaces используются для зависимостей между пакетами (см. `routing.go`).

### mcoreclient Library
Все workers используют `mcoreclient` для коммуникации с mcore:
- `ListeningTasks()` - фоновый процесс для опроса задач
- `AddMsg()` - отправка сообщений в mcore
- `GetTask()` / `ReportTask()` - работа с задачами

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