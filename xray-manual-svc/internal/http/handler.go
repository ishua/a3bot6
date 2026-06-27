package http

import (
    "encoding/json"
    "net/http"
    "xray-manual-svc/internal"
)

type Handler struct {
    manager    *internal.ProxyManager
    appVersion string
}

func NewHandler(manager *internal.ProxyManager, appVersion string) *Handler {
    return &Handler{manager: manager, appVersion: appVersion}
}

type response struct {
    Data  any     `json:"data"`
    Error *string `json:"error"`
}

func writeOK(w http.ResponseWriter, data any) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response{Data: data})
}

func writeError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    msg := err.Error()
    json.NewEncoder(w).Encode(response{Error: &msg})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    _, err := h.manager.Status()
    if err != nil {
        writeError(w, err)
        return
    }
    writeOK(w, map[string]string{"status": "ok", "version": h.appVersion})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    writeOK(w, map[string]any{"tags": h.manager.List()})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
    status, err := h.manager.Status()
    if err != nil {
        writeError(w, err)
        return
    }
    writeOK(w, map[string]string{
        "override":         status.Override,
        "principle_target": status.PrincipleTarget,
    })
}

func (h *Handler) Use(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Tag string `json:"tag"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, err)
        return
    }
    if err := h.manager.Use(req.Tag); err != nil {
        writeError(w, err)
        return
    }
    writeOK(w, map[string]string{"tag": req.Tag})
}

func (h *Handler) Auto(w http.ResponseWriter, r *http.Request) {
    if err := h.manager.Auto(); err != nil {
        writeError(w, err)
        return
    }
    writeOK(w, nil)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
    result, err := h.manager.Ping()
    if err != nil {
        writeError(w, err)
        return
    }
    writeOK(w, map[string]any{
        "ip":         result.IP,
        "latency_ms": result.Latency.Milliseconds(),
    })
}