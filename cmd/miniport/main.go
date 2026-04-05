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

func main() {
	host := envOr("MINIPORT_HOST", "127.0.0.1")
	port := envOr("MINIPORT_PORT", "8092")
	logRequests := envOr("MINIPORT_LOG_REQUESTS", "false") == "true"
	logTailLines, _ := strconv.Atoi(envOr("MINIPORT_LOG_TAIL_LINES", "100"))
	if logTailLines <= 0 {
		logTailLines = 100
	}
	statsInterval, _ := strconv.Atoi(envOr("MINIPORT_STATS_INTERVAL", "15"))
	if statsInterval <= 0 {
		statsInterval = 15
	}

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

	serviceNames := parseServices(envOr("MINIPORT_SERVICES", ""))

	collector := stats.NewCollector(dockerSvc, time.Duration(statsInterval)*time.Second, 60, serviceNames)

	h := handler.New(dockerSvc, logTailLines, collector, serviceNames)

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
	if len(serviceNames) > 0 {
		mux.HandleFunc("GET /services", h.ServiceTable)
		mux.HandleFunc("POST /services/{name}/start", h.ServiceStart)
		mux.HandleFunc("POST /services/{name}/stop", h.ServiceStop)
		mux.HandleFunc("POST /services/{name}/restart", h.ServiceRestart)
		mux.HandleFunc("GET /services/{name}/logs", h.ServiceLogs)
	}
	mux.HandleFunc("POST /prune/containers", h.PruneContainers)
	mux.HandleFunc("POST /prune/images", h.PruneImages)
	mux.HandleFunc("POST /prune/volumes", h.PruneVolumes)
	mux.HandleFunc("POST /prune/networks", h.PruneNetworks)

	addr := fmt.Sprintf("%s:%s", host, port)
	var h2 http.Handler = mux
	if logRequests {
		h2 = handler.Logger(h2)
	}
	h2 = handler.Recover(h2)
	srv := &http.Server{
		Addr:    addr,
		Handler: h2,
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
