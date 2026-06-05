package main

import (
	"log"
	"net/http"

	"github.com/a3bot6/a3b-webui/internal/api"
	"github.com/a3bot6/a3b-webui/internal/auth"
	"github.com/a3bot6/a3b-webui/internal/config"
	"github.com/a3bot6/a3b-webui/internal/handler"
	"github.com/a3bot6/a3b-webui/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// API client
	apiClient := api.New(cfg.API.URL, cfg.API.Secret)

	// Session manager
	sessionManager := auth.NewSessionManager(cfg.Auth.SessionSecret)

	// Handlers
	authH := handler.NewAuthHandler(cfg.Auth.Login, cfg.Auth.Password, sessionManager)
	dashboardH := handler.NewDashboardHandler(apiClient)
	vpnH := handler.NewVPNHandler(apiClient)

	// Router
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("GET /login", authH.LoginPage)
	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("POST /logout", authH.Logout)

	// Dashboard
	mux.HandleFunc("GET /", dashboardH.Dashboard)

	// VPN actions
	mux.HandleFunc("POST /use/{tag}", vpnH.Use)
	mux.HandleFunc("POST /auto", vpnH.Auto)
	mux.HandleFunc("GET /ping", vpnH.Ping)

	// Wrap with auth middleware
	handler := middleware.AuthMiddleware(mux, sessionManager)

	addr := cfg.Server.Addr
	log.Printf("starting a3b-webui on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
