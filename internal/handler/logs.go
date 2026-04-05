package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) Logs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	lines := h.LogTailLines
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lines = n
		}
	}
	logs, err := h.docker.Logs(r.Context(), id, lines)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	renderPartial(w, "logs-panel.html", map[string]any{"ContainerID": id, "Logs": logs})
}
