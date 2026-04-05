package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func setPruneTrigger(w http.ResponseWriter, msg string) {
	trigger := map[string]any{
		"refresh-containers": "",
		"showToast":          map[string]string{"msg": msg, "type": "success"},
	}
	b, _ := json.Marshal(trigger)
	w.Header().Set("HX-Trigger", string(b))
}

func (h *Handler) PruneContainers(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneContainers(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setPruneTrigger(w, fmt.Sprintf("Containers pruned — reclaimed %s", FormatBytes(freed)))
	w.WriteHeader(200)
}

func (h *Handler) PruneImages(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneImages(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setPruneTrigger(w, fmt.Sprintf("Images pruned — reclaimed %s", FormatBytes(freed)))
	w.WriteHeader(200)
}

func (h *Handler) PruneVolumes(w http.ResponseWriter, r *http.Request) {
	freed, err := h.docker.PruneVolumes(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setPruneTrigger(w, fmt.Sprintf("Volumes pruned — reclaimed %s", FormatBytes(freed)))
	w.WriteHeader(200)
}

func (h *Handler) PruneNetworks(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.PruneNetworks(r.Context()); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setPruneTrigger(w, "Networks pruned")
	w.WriteHeader(200)
}

// CapPct caps a percentage value at 100 (CPU% can exceed 100 on multi-core).
func CapPct(v float64) float64 {
	if v > 100 {
		return 100
	}
	return v
}

// FormatMB formats bytes as a human-readable MB value.
func FormatMB(b uint64) string {
	mb := float64(b) / 1024 / 1024
	if mb < 100 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	return fmt.Sprintf("%.0f MB", mb)
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
