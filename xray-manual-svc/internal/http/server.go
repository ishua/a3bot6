package http

import (
    "net/http"
)

func NewServer(addr string, handler *Handler, secrets []string) *http.Server {
    mux := http.NewServeMux()

    mux.HandleFunc("/health", handler.Health)
    mux.HandleFunc("/list", handler.List)
    mux.HandleFunc("/status", handler.Status)
    mux.HandleFunc("/use", handler.Use)
    mux.HandleFunc("/auto", handler.Auto)
    mux.HandleFunc("/ping", handler.Ping)

    return &http.Server{
        Addr:    addr,
        Handler: MiddleAuth(mux, secrets),
    }
}
