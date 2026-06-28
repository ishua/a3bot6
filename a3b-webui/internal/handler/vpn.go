package handler

import (
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

// Use switches VPN to a specific tag (POST /use/{tag}).
func (h *VPNHandler) Use(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if tag == "" {
		http.Error(w, "Missing tag", http.StatusBadRequest)
		return
	}

	if err := h.apiCli.Use(tag); err != nil {
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
		component := templates.PingError(err.Error())
		component.Render(r.Context(), w)
		return
	}

	component := templates.PingResult(result.IP, result.LatencyMs)
	component.Render(r.Context(), w)
}
