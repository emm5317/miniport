package handler

import "github.com/gofiber/fiber/v3"

func (h *Handler) Stats(c fiber.Ctx) error {
	id := c.Params("id")
	stats, err := h.docker.Stats(c.Context(), id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Render("partials/stats-modal", fiber.Map{"ContainerID": id, "Stats": stats})
}
