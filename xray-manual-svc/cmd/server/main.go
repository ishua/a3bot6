package main

import (
    "log"
    "xray-manual-svc/internal"
    "xray-manual-svc/internal/app/config"
    internalhttp "xray-manual-svc/internal/http"
)

var appVersion = "dev"

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    manager, err := internal.Bootstrap(cfg)
    if err != nil {
        log.Fatalf("failed to bootstrap: %v", err)
    }

    handler := internalhttp.NewHandler(manager, appVersion)
    server := internalhttp.NewServer(cfg.Server.Addr, handler, cfg.Auth.Secrets)

    log.Printf("starting server on %s", cfg.Server.Addr)
    log.Printf("version: %s", appVersion)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
