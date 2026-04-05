package handler

import (
	"fmt"
	"net/http"
)

func (h *Handler) PruneContainers(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneContainers(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Reclaimed %s", FormatBytes(freed))
}

func (h *Handler) PruneImages(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneImages(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Reclaimed %s", FormatBytes(freed))
}

func (h *Handler) PruneVolumes(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneVolumes(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Reclaimed %s", FormatBytes(freed))
}

func (h *Handler) PruneNetworks(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.PruneNetworks(r.Context()); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Networks pruned"))
}

func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
