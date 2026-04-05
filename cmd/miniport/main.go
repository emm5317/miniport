package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"github.com/emm5317/miniport/internal/docker"
	"github.com/emm5317/miniport/internal/handler"
	"github.com/emm5317/miniport/web"
)

func main() {
	host := envOr("MINIPORT_HOST", "127.0.0.1")
	port := envOr("MINIPORT_PORT", "8092")

	dockerSvc, err := docker.NewService()
	if err != nil {
		log.Fatalf("Docker: %v", err)
	}

	engine := html.NewFileSystem(http.FS(web.Templates), ".html")
	engine.AddFuncMap(map[string]any{
		"formatBytes": handler.FormatBytes,
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(logger.New())
	app.Use(recoverer.New())

	h := handler.New(dockerSvc)

	app.Get("/healthz", h.Healthz)
	app.Get("/", h.Index)
	app.Get("/containers", h.ContainerTable)
	app.Post("/containers/:id/start", h.Start)
	app.Post("/containers/:id/stop", h.Stop)
	app.Post("/containers/:id/restart", h.Restart)
	app.Delete("/containers/:id", h.Remove)
	app.Get("/containers/:id/logs", h.Logs)
	app.Get("/containers/:id/stats", h.Stats)
	app.Post("/prune/containers", h.PruneContainers)
	app.Post("/prune/images", h.PruneImages)
	app.Post("/prune/volumes", h.PruneVolumes)
	app.Post("/prune/networks", h.PruneNetworks)
	addr := fmt.Sprintf("%s:%s", host, port)
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
