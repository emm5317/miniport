# MiniPort

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-SDK-2496ED?logo=docker&logoColor=white)](https://docs.docker.com/engine/api/sdk/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A lightweight Docker and systemd dashboard. Single Go binary, no database, no container required. Built for resource-constrained servers.

## Features

### Containers
- **Dashboard** — live container table with search, state filters (All/Running/Stopped), and sortable columns
- **Inline stats** — CPU and memory bars rendered per container row, updated via background collector
- **Sparkline charts** — SVG CPU and memory trend lines from ring buffer history
- **Actions** — start, stop, restart, remove with toast notifications
- **Logs** — configurable tail lines (50/100/500/1000), in-log search with highlighting, live streaming via 2s polling, pause/resume on scroll, copy-to-clipboard
- **Stats panel** — live CPU, memory, network, and disk I/O with history sparklines
- **Inspect panel** — container config, environment (masked by default), ports, networks, mounts, and labels in CSS-only tabbed interface
- **Prune** — clean up containers, images, volumes, and networks

### Host
- **System metrics** — live CPU, memory, disk, and network stats from the host (Linux only, via `/proc`)
- **Uptime** — displayed in the host strip above the container summary

### Systemd Services
- **Service monitoring** — track configured systemd services alongside Docker containers
- **Live status** — active/inactive/failed state with CPU% and memory usage
- **Actions** — start, stop, restart with toast notifications
- **Logs** — journal output with the same search, streaming, and controls as container logs
- **Zero dependencies** — uses `systemctl` and `journalctl` via `os/exec`, no D-Bus library

### General
- **Toast notifications** — success/error feedback on all actions
- **Health endpoint** — `GET /healthz` for uptime monitors
- **Auto-refresh** — container and service tables poll every 10s, stats every 5s

## Stack

| Layer | Choice |
|-------|--------|
| Backend | Go stdlib `net/http` (zero framework dependencies) |
| Docker | [Official SDK](https://pkg.go.dev/github.com/docker/docker/client) |
| Systemd | `systemctl` / `journalctl` via `os/exec` (Linux only) |
| Frontend | HTMX (only external JS dependency) |
| Templates | Go `html/template` via `go:embed` |
| CSS | Embedded stylesheet, no CDN, no framework |

No JavaScript build step. No database. No framework. All state comes from the Docker daemon and systemd.

## Requirements

- Go 1.22+
- Docker daemon running (accessible via `/var/run/docker.sock`)
- Linux for host metrics and systemd service monitoring (graceful no-op on other platforms)

## Quick Start

```bash
git clone https://github.com/emm5317/miniport.git
cd miniport
go build -ldflags="-s -w" -o miniport ./cmd/miniport
./miniport
```

Listens on `127.0.0.1:8092` by default. Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIPORT_HOST` | `127.0.0.1` | Bind address |
| `MINIPORT_PORT` | `8092` | Bind port |
| `MINIPORT_LOG_TAIL_LINES` | `100` | Default log lines per container/service |
| `MINIPORT_LOG_REQUESTS` | `false` | Enable HTTP request logging |
| `MINIPORT_STATS_INTERVAL` | `15` | Background stats collection interval (seconds) |
| `MINIPORT_SERVICES` | *(empty)* | Comma-separated systemd service names to monitor |

## Deploy with systemd

```ini
[Unit]
Description=MiniPort Docker Dashboard
After=network.target docker.service

[Service]
Type=simple
ExecStart=/opt/miniport/miniport
EnvironmentFile=/opt/miniport/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Example `/opt/miniport/.env`:
```
MINIPORT_SERVICES=voicetask,caddy,postgresql
```

Put behind a reverse proxy (Caddy, nginx) with authentication — the Docker socket is root-equivalent access.

## Resource Usage

- **Binary size:** ~11 MB (stripped)
- **Idle RAM:** <20 MB
- **Direct dependencies:** 1 (Docker SDK)
- **No streaming connections** — stats are polled snapshots, not persistent goroutines

## License

MIT
