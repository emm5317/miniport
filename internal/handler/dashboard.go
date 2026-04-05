package handler

import (
	"net/http"

	"github.com/emm5317/miniport/internal/docker"
)

type Handler struct {
	docker       *docker.Service
	LogTailLines int
}

func New(d *docker.Service, logTailLines int) *Handler {
	return &Handler{docker: d, LogTailLines: logTailLines}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.List(r.Context())
	if err != nil {
		httpError(w, "Failed to list containers: "+err.Error(), 500)
		return
	}
	summary := docker.Summarize(containers)
	renderPage(w, "pages/index", map[string]any{
		"Containers": containers,
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
	renderPartial(w, "partials/container-table.html", map[string]any{
		"Containers": containers,
		"Summary":    summary,
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
