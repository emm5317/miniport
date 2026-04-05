package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emm5317/miniport/internal/systemd"
)

func (h *Handler) ServiceLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	lines := h.LogTailLines
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lines = n
		}
	}

	since := r.URL.Query().Get("since")
	stream := r.URL.Query().Get("stream") == "true"

	if stream && since != "" {
		logs, err := systemd.LogsSince(r.Context(), name, since)
		if err != nil {
			httpError(w, err.Error(), 500)
			return
		}
		w.Header().Set("X-Log-Timestamp", time.Now().UTC().Format(time.RFC3339Nano))
		if logs == "" {
			w.WriteHeader(204)
			return
		}
		renderPartial(w, "logs-lines.html", map[string]any{"Lines": colorizeLogs(logs)})
		return
	}

	logs, err := systemd.Logs(r.Context(), name, lines)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	renderPartial(w, "logs-panel.html", map[string]any{
		"ContainerID": name,
		"Logs":        colorizeLogs(logs),
		"Lines":       lines,
		"BasePath":    "/services/" + name,
	})
}
