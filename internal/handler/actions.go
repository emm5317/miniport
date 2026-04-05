package handler

import "github.com/gofiber/fiber/v3"

func (h *Handler) Start(c fiber.Ctx) error {
	if err := h.docker.Start(c.Context(), c.Params("id")); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("HX-Trigger", "refresh-containers")
	return c.SendStatus(200)
}

func (h *Handler) Stop(c fiber.Ctx) error {
	if err := h.docker.Stop(c.Context(), c.Params("id")); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("HX-Trigger", "refresh-containers")
	return c.SendStatus(200)
}

func (h *Handler) Restart(c fiber.Ctx) error {
	if err := h.docker.Restart(c.Context(), c.Params("id")); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("HX-Trigger", "refresh-containers")
	return c.SendStatus(200)
}

func (h *Handler) Remove(c fiber.Ctx) error {
	if err := h.docker.Remove(c.Context(), c.Params("id")); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("HX-Trigger", "refresh-containers")
	return c.SendStatus(200)
}
