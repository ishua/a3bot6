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

// Dashboard renders the dashboard page (GET /).
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	status, err := h.apiCli.GetStatus()
	if err != nil {
		component := templates.ErrorPage("xray-manual-svc is not available")
		component.Render(r.Context(), w)
		return
	}

	tags, err := h.apiCli.GetList()
	if err != nil {
		component := templates.ErrorPage("xray-manual-svc is not available")
		component.Render(r.Context(), w)
		return
	}

	activeTag := status.Override

	// HTMX request — return only the fragment (for Refresh button)
	if r.Header.Get("HX-Request") == "true" {
		component := templates.DashboardFragment(status, tags, activeTag)
		component.Render(r.Context(), w)
		return
	}

	component := templates.DashboardPage(status, tags, activeTag)
	component.Render(r.Context(), w)
}
