package handler

import (
	"fmt"
	"net/http"
)

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stats, err := h.docker.Stats(r.Context(), id)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	renderPartial(w, "stats-modal.html", map[string]any{"ContainerID": id, "Stats": stats})
}

func (h *Handler) InlineStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stats, err := h.docker.Stats(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<td class="py-3 pr-4 text-gray-600 text-xs">—</td><td class="py-3 pr-4 text-gray-600 text-xs">—</td>`)
		return
	}
	renderPartial(w, "inline-stats.html", stats)
}
