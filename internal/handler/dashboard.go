package handler

import (
	"net/http"
	"sync"

	"github.com/emm5317/miniport/internal/docker"
)

// ContainerRow pairs a container with its live stats (nil if not running).
type ContainerRow struct {
	docker.ContainerInfo
	Stats *docker.StatsSnapshot
}

type Handler struct {
	docker       *docker.Service
	LogTailLines int
}

func New(d *docker.Service, logTailLines int) *Handler {
	return &Handler{docker: d, LogTailLines: logTailLines}
}

// buildRows fetches stats concurrently for running containers.
func (h *Handler) buildRows(r *http.Request, containers []docker.ContainerInfo) []ContainerRow {
	rows := make([]ContainerRow, len(containers))
	var wg sync.WaitGroup
	for i, c := range containers {
		rows[i] = ContainerRow{ContainerInfo: c}
		if c.State == "running" {
			wg.Add(1)
			go func(idx int, id string) {
				defer wg.Done()
				stats, err := h.docker.Stats(r.Context(), id)
				if err == nil {
					rows[idx].Stats = stats
				}
			}(i, c.ID)
		}
	}
	wg.Wait()
	return rows
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.List(r.Context())
	if err != nil {
		httpError(w, "Failed to list containers: "+err.Error(), 500)
		return
	}
	summary := docker.Summarize(containers)
	rows := h.buildRows(r, containers)
	renderPage(w, "pages/index", map[string]any{
		"Containers": rows,
		"Summary":    summary,
	})
}

func (h *Handler) ContainerTable(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.List(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	summary := docker.Summarize(containers)
	rows := h.buildRows(r, containers)
	renderPartial(w, "container-table.html", map[string]any{
		"Containers": rows,
		"Summary":    summary,
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
