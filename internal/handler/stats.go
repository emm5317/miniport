package handler

import "net/http"

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stats, err := h.docker.Stats(r.Context(), id)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	history := h.collector.ContainerHistory(id)
	renderPartial(w, "stats-modal.html", map[string]any{
		"ContainerID": id,
		"Stats":       stats,
		"History":     history,
	})
}
