# MiniPort

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-SDK-2496ED?logo=docker&logoColor=white)](https://docs.docker.com/engine/api/sdk/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A lightweight Docker dashboard. Single Go binary, no database, no container required. Built for resource-constrained servers.

## Features

- **Dashboard** — live container table with summary strip (total/running/stopped/unhealthy counts + SVG status bar)
- **Actions** — start, stop, restart, remove with confirmation
- **Logs** — configurable tail lines with copy-to-clipboard, properly demuxed stdout/stderr
- **Stats** — live CPU, memory, network, and disk I/O (polled every 5s)
- **Prune** — clean up containers, images, volumes, and networks
- **Health endpoint** — `GET /healthz` for uptime monitors

## Stack

| Layer | Choice |
|-------|--------|
| Backend | Go stdlib `net/http` (zero framework dependencies) |
| Docker | [Official SDK](https://pkg.go.dev/github.com/docker/docker/client) |
| Frontend | HTMX (only external dependency) |
| Templates | Go `html/template` via `go:embed` |
| CSS | Embedded stylesheet (no CDN) |

No JavaScript build step. No database. No framework. All state comes from the Docker daemon.

## Requirements

- Go 1.22+
- Docker daemon running (accessible via `/var/run/docker.sock`)

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
| `MINIPORT_LOG_TAIL_LINES` | `100` | Number of log lines to fetch per container |
| `MINIPORT_LOG_REQUESTS` | `false` | Enable HTTP request logging |

## Deploy with systemd

```ini
[Unit]
Description=MiniPort Docker Dashboard
After=network.target docker.service

[Service]
Type=simple
ExecStart=/opt/miniport/miniport
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Put behind a reverse proxy (Caddy, nginx) with authentication — the Docker socket is root-equivalent access.

## Resource Usage

- **Binary size:** ~11 MB (stripped)
- **Idle RAM:** <20 MB
- **Direct dependencies:** 1 (Docker SDK)
- **No streaming connections** — stats are polled snapshots, not persistent goroutines

## License

MIT
