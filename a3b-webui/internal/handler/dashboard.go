package handler

import (
	"net/http"

	"github.com/a3bot6/a3b-webui/internal/api"
	"github.com/a3bot6/a3b-webui/templates"
)

// DashboardHandler handles the main dashboard page.
type DashboardHandler struct {
	apiCli api.ClientInterface
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(apiCli api.ClientInterface) *DashboardHandler {
	return &DashboardHandler{apiCli: apiCli}
}

// Dashboard renders the full dashboard page (GET /).
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
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
	component := templates.DashboardPage(status, tags, activeTag)
	component.Render(r.Context(), w)
}
