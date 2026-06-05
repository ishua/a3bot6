package handler

import (
	"encoding/json"
	"net/http"

	"github.com/a3bot6/a3b-webui/internal/api"
	"github.com/a3bot6/a3b-webui/templates"
)

// VPNHandler handles VPN actions (use, auto, ping).
type VPNHandler struct {
	apiCli api.ClientInterface
}

// NewVPNHandler creates a new VPNHandler.
func NewVPNHandler(apiCli api.ClientInterface) *VPNHandler {
	return &VPNHandler{apiCli: apiCli}
}

// Use switches VPN to a specific tag (POST /use).
func (h *VPNHandler) Use(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tag string `json:"tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.apiCli.Use(req.Tag); err != nil {
		http.Error(w, "Failed to switch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Re-render full dashboard fragment
	status, err := h.apiCli.GetStatus()
	if err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}
	tags, err := h.apiCli.GetList()
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}
	activeTag := status.Override
	component := templates.DashboardFragment(status, tags, activeTag)
	component.Render(r.Context(), w)
}

// Auto enables auto mode (POST /auto).
func (h *VPNHandler) Auto(w http.ResponseWriter, r *http.Request) {
	if err := h.apiCli.Auto(); err != nil {
		http.Error(w, "Failed to enable auto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Re-render full dashboard fragment
	status, err := h.apiCli.GetStatus()
	if err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}
	tags, err := h.apiCli.GetList()
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}
	activeTag := status.Override
	component := templates.DashboardFragment(status, tags, activeTag)
	component.Render(r.Context(), w)
}

// Ping pings the current VPN (GET /ping) — returns HTMX fragment.
func (h *VPNHandler) Ping(w http.ResponseWriter, r *http.Request) {
	result, err := h.apiCli.Ping()
	if err != nil {
		http.Error(w, "Ping failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	component := templates.PingResult(result.IP, result.LatencyMs)
	component.Render(r.Context(), w)
}
