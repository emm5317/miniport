package handler

import (
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) Logs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	lines := h.LogTailLines
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lines = n
		}
	}

	since := r.URL.Query().Get("since")
	stream := r.URL.Query().Get("stream") == "true"

	if stream && since != "" {
		logs, err := h.docker.LogsSince(r.Context(), id, 0, since)
		if err != nil {
			httpError(w, err.Error(), 500)
			return
		}
		// Return the current server timestamp for the next poll
		w.Header().Set("X-Log-Timestamp", time.Now().UTC().Format(time.RFC3339Nano))
		if logs == "" {
			w.WriteHeader(204)
			return
		}
		renderPartial(w, "logs-lines.html", map[string]any{"Lines": logs})
		return
	}

	logs, err := h.docker.Logs(r.Context(), id, lines)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	renderPartial(w, "logs-panel.html", map[string]any{
		"ContainerID": id,
		"Logs":        logs,
		"Lines":       lines,
	})
}
