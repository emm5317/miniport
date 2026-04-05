package handler

import "github.com/gofiber/fiber/v3"

func (h *Handler) Logs(c fiber.Ctx) error {
	id := c.Params("id")
	lines := fiber.Query(c, "lines", 200)
	logs, err := h.docker.Logs(c.Context(), id, lines)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Render("partials/logs-panel", fiber.Map{"ContainerID": id, "Logs": logs})
}
