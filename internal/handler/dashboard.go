package handler

import (
	"net/http"

	"github.com/emm5317/miniport/internal/docker"
	"github.com/emm5317/miniport/internal/stats"
)

// ContainerRow pairs a container with its live stats (nil if not running).
type ContainerRow struct {
	docker.ContainerInfo
	Stats   *docker.StatsSnapshot
	History []docker.StatsSnapshot
}

type Handler struct {
	docker       *docker.Service
	LogTailLines int
	collector    *stats.Collector
}

func New(d *docker.Service, logTailLines int, c *stats.Collector) *Handler {
	return &Handler{docker: d, LogTailLines: logTailLines, collector: c}
}

// buildRows reads latest stats from the collector instead of fetching live.
func (h *Handler) buildRows(containers []docker.ContainerInfo) []ContainerRow {
	latest := h.collector.AllLatest()
	rows := make([]ContainerRow, len(containers))
	for i, c := range containers {
		rows[i] = ContainerRow{ContainerInfo: c}
		if s, ok := latest[c.ID]; ok {
			rows[i].Stats = s
		}
		rows[i].History = h.collector.ContainerHistory(c.ID)
	}
	return rows
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.List(r.Context())
	if err != nil {
		httpError(w, "Failed to list containers: "+err.Error(), 500)
		return
	}
	summary := docker.Summarize(containers)
	rows := h.buildRows(containers)
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
	rows := h.buildRows(containers)
	renderPartial(w, "container-table.html", map[string]any{
		"Containers": rows,
		"Summary":    summary,
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
