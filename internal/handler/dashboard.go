package handler

import "github.com/emm5317/miniport/internal/docker"
import "github.com/gofiber/fiber/v3"

type Handler struct {
	docker *docker.Service
}

func New(d *docker.Service) *Handler {
	return &Handler{docker: d}
}

func (h *Handler) Index(c fiber.Ctx) error {
	containers, err := h.docker.List(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to list containers: " + err.Error())
	}
	return c.Render("pages/index", fiber.Map{"Containers": containers}, "layouts/base")
}

func (h *Handler) ContainerTable(c fiber.Ctx) error {
	containers, err := h.docker.List(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Render("partials/container-table", fiber.Map{"Containers": containers})
}

func (h *Handler) Healthz(c fiber.Ctx) error {
	return c.SendString("ok")
}
