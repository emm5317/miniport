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

// ServiceRow holds a systemd service's current state for display.
type ServiceRow struct {
	Name        string
	Description string
	ActiveState string
	SubState    string
	MemCurrent  uint64
	CPUPercent  float64
	StartedAt   string
	NRestarts   int
	UnitEnabled string
}

type Handler struct {
	docker       *docker.Service
	LogTailLines int
	collector    *stats.Collector
	serviceNames []string
}

func New(d *docker.Service, logTailLines int, c *stats.Collector, serviceNames []string) *Handler {
	return &Handler{docker: d, LogTailLines: logTailLines, collector: c, serviceNames: serviceNames}
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

func (h *Handler) buildServiceRows() []ServiceRow {
	if len(h.serviceNames) == 0 {
		return nil
	}
	latest := h.collector.AllServiceLatest()
	rows := make([]ServiceRow, 0, len(h.serviceNames))
	for _, name := range h.serviceNames {
		s, ok := latest[name]
		if !ok {
			rows = append(rows, ServiceRow{Name: name, ActiveState: "unknown"})
			continue
		}
		rows = append(rows, ServiceRow{
			Name:        s.Name,
			Description: s.Description,
			ActiveState: s.ActiveState,
			SubState:    s.SubState,
			MemCurrent:  s.MemCurrent,
			CPUPercent:  s.CPUPercent,
			StartedAt:   s.StartedAt,
			NRestarts:   s.NRestarts,
			UnitEnabled: s.UnitEnabled,
		})
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
	data := map[string]any{
		"Containers": rows,
		"Summary":    summary,
		"Host":       h.collector.HostLatest(),
	}
	if svcs := h.buildServiceRows(); len(svcs) > 0 {
		data["Services"] = svcs
	}
	renderPage(w, "pages/index", data)
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
		"Host":       h.collector.HostLatest(),
	})
}

func (h *Handler) ServiceTable(w http.ResponseWriter, r *http.Request) {
	svcs := h.buildServiceRows()
	renderPartial(w, "service-table.html", map[string]any{
		"Services": svcs,
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
