package main

import (
    "fmt"
    "log"
    "os"
    "xray-manual-svc/internal"
    "xray-manual-svc/internal/app/config"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    manager, err := internal.Bootstrap(cfg)
    if err != nil {
        log.Fatalf("failed to bootstrap: %v", err)
    }

    switch os.Args[1] {
    case "list":
        for _, t := range manager.List() {
            fmt.Println(t)
        }

    case "status":
        status, err := manager.Status()
        if err != nil {
            log.Fatalf("failed to get status: %v", err)
        }
        fmt.Printf("override:         %s\n", status.Override)
        fmt.Printf("principle_target: %s\n", status.PrincipleTarget)

    case "use":
        if len(os.Args) < 3 {
            log.Fatalf("usage: use <tag>")
        }
        if err := manager.Use(os.Args[2]); err != nil {
            log.Fatalf("failed to set target: %v", err)
        }
        fmt.Printf("target set: %s\n", os.Args[2])

    case "auto":
        if err := manager.Auto(); err != nil {
            log.Fatalf("failed to reset target: %v", err)
        }
        fmt.Println("target reset to auto")

    case "ping":
        result, err := manager.Ping()
        if err != nil {
            log.Fatalf("failed to ping: %v", err)
        }
        fmt.Printf("ip:      %s\n", result.IP)
        fmt.Printf("latency: %s\n", result.Latency)

    default:
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("usage:")
    fmt.Println("  cli list")
    fmt.Println("  cli status")
    fmt.Println("  cli use <tag>")
    fmt.Println("  cli auto")
    fmt.Println("  cli ping")
}