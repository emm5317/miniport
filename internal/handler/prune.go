package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func (h *Handler) PruneContainers(c fiber.Ctx) error {
	freed, err := h.docker.PruneContainers(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString(fmt.Sprintf("Reclaimed %s", FormatBytes(freed)))
}

func (h *Handler) PruneImages(c fiber.Ctx) error {
	freed, err := h.docker.PruneImages(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString(fmt.Sprintf("Reclaimed %s", FormatBytes(freed)))
}

func (h *Handler) PruneVolumes(c fiber.Ctx) error {
	freed, err := h.docker.PruneVolumes(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString(fmt.Sprintf("Reclaimed %s", FormatBytes(freed)))
}

func (h *Handler) PruneNetworks(c fiber.Ctx) error {
	if err := h.docker.PruneNetworks(c.Context()); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString("Networks pruned")
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
