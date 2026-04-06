package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/emm5317/miniport/internal/docker"
	"github.com/emm5317/miniport/internal/handler"
	"github.com/emm5317/miniport/internal/stats"
	"github.com/emm5317/miniport/web"
)

type config struct {
	Host          string
	Port          string
	LogRequests   bool
	LogTailLines  int
	StatsInterval int
	Services      []string
	AuthFile      string
}

func loadConfig() config {
	cfg := config{
		Host:          envOr("MINIPORT_HOST", "127.0.0.1"),
		Port:          envOr("MINIPORT_PORT", "8092"),
		LogRequests:   envOr("MINIPORT_LOG_REQUESTS", "false") == "true",
		LogTailLines:  100,
		StatsInterval: 15,
		Services:      parseServices(envOr("MINIPORT_SERVICES", "")),
		AuthFile:      envOr("MINIPORT_AUTH", ""),
	}
	if v, err := strconv.Atoi(envOr("MINIPORT_LOG_TAIL_LINES", "100")); err == nil && v > 0 {
		cfg.LogTailLines = v
	}
	if v, err := strconv.Atoi(envOr("MINIPORT_STATS_INTERVAL", "15")); err == nil && v > 0 {
		cfg.StatsInterval = v
	}
	return cfg
}

func main() {
	cfg := loadConfig()

	dockerSvc, err := docker.NewService()
	if err != nil {
		log.Fatalf("Docker: %v", err)
	}

	tmplSub, _ := fs.Sub(web.Templates, "templates")
	handler.InitTemplates(tmplSub, template.FuncMap{
		"formatBytes": handler.FormatBytes,
		"formatMB":    handler.FormatMB,
		"pct": func(part, total int) int {
			if total == 0 {
				return 0
			}
			return part * 100 / total
		},
		"addPct": func(a, b, total int) int {
			if total == 0 {
				return 0
			}
			return (a + b) * 100 / total
		},
		"capPct":       handler.CapPct,
		"sparkline":    handler.Sparkline,
		"sparklineMem": handler.SparklineMem,
	})

	staticSub, _ := fs.Sub(web.Static, "static")

	collector := stats.NewCollector(dockerSvc, time.Duration(cfg.StatsInterval)*time.Second, 60, cfg.Services)

	h := handler.New(dockerSvc, cfg.LogTailLines, collector, cfg.Services)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /{$}", h.Index)
	mux.HandleFunc("GET /containers", h.ContainerTable)
	mux.HandleFunc("POST /containers/{id}/start", h.Start)
	mux.HandleFunc("POST /containers/{id}/stop", h.Stop)
	mux.HandleFunc("POST /containers/{id}/restart", h.Restart)
	mux.HandleFunc("DELETE /containers/{id}", h.Remove)
	mux.HandleFunc("GET /containers/{id}/logs", h.Logs)
	mux.HandleFunc("GET /containers/{id}/stats", h.Stats)
	mux.HandleFunc("GET /containers/{id}/inspect", h.Inspect)
	mux.HandleFunc("POST /containers/{id}/recreate", h.Recreate)
	if len(cfg.Services) > 0 {
		mux.HandleFunc("GET /services", h.ServiceTable)
		mux.HandleFunc("POST /services/{name}/start", h.ServiceStart)
		mux.HandleFunc("POST /services/{name}/stop", h.ServiceStop)
		mux.HandleFunc("POST /services/{name}/restart", h.ServiceRestart)
		mux.HandleFunc("GET /services/{name}/logs", h.ServiceLogs)
	}
	mux.HandleFunc("GET /images", h.ImageList)
	mux.HandleFunc("POST /images/pull", h.ImagePull)
	mux.HandleFunc("DELETE /images/{id}", h.ImageRemove)
	mux.HandleFunc("POST /prune/containers", h.PruneContainers)
	mux.HandleFunc("POST /prune/images", h.PruneImages)
	mux.HandleFunc("POST /prune/volumes", h.PruneVolumes)
	mux.HandleFunc("POST /prune/networks", h.PruneNetworks)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var h2 http.Handler = mux
	h2 = handler.CheckOrigin(h2)
	if cfg.LogRequests {
		h2 = handler.Logger(h2)
	}
	h2 = handler.Recover(h2)

	// Optional basic auth via auth file.
	if cfg.AuthFile != "" {
		creds, err := handler.LoadHtpasswd(cfg.AuthFile)
		if err != nil {
			log.Fatalf("Auth: %v", err)
		}
		h2 = handler.BasicAuth(creds, h2)
		log.Printf("Basic auth enabled (%d user(s) from %s)", len(creds), cfg.AuthFile)
	}

	h2 = handler.SecureHeaders(h2)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      180 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go collector.Start(ctx)

	go func() {
		log.Printf("Listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseServices(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}
