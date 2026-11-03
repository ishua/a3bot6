# Root Makefile for the a3bot6 Go workspace (see go.work).
# There is no root Go module: `go build ./...` from the repo root does not
# work, so targets below run per module. mcore needs CGO_ENABLED=1 (mattn/go-sqlite3);
# the flag is harmless for the pure-Go modules. ytd2feed is plain Python —
# it is not covered by the Go build/test loop
# (run: pip install -r ytd2feed/requirements.txt && python ytd2feed/main.py).
# mcore/Makefile has local run helpers: `make -C mcore run` builds and starts mcore.

MODULES := mcore tbot notes

.PHONY: help build test

help: ## list targets and modules
	@echo "Targets: help (default) | build (Go modules only, sequential, fail fast) | test (go test ./... per Go module)"
	@echo "Modules:"
	@echo "  mcore    REST API core: command parsing, task queue, SQLite state (needs CGO)"
	@echo "  tbot     Telegram gateway: TG -> mcore, mcore -> TG"
	@echo "  notes    notes worker: git-based notes, /note task parsing"
	@echo "  ytd2feed Python worker (yt-dlp -> RSS): not covered by build/test"

build: ## CGO_ENABLED=1 go build ./... in each Go module, sequential, fail fast
	set -e; \
	for m in $(MODULES); do \
		echo "== build $$m =="; \
		(cd $$m && CGO_ENABLED=1 go build ./...); \
	done; \
	echo "== all Go modules built =="

test: ## go test ./... in each Go module, sequential, fail fast
	set -e; \
	for m in $(MODULES); do \
		echo "== test $$m =="; \
		(cd $$m && CGO_ENABLED=1 go test ./...); \
	done; \
	echo "== all Go modules tested =="
