# MiniPort

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-SDK-2496ED?logo=docker&logoColor=white)](https://docs.docker.com/engine/api/sdk/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A lightweight Docker dashboard. Single Go binary, no database, no container required. Built for resource-constrained servers.

## Features

- **Dashboard** — live container table with name, image, state, status, ports
- **Actions** — start, stop, restart, remove with confirmation
- **Logs** — last 200 lines with copy-to-clipboard, properly demuxed stdout/stderr
- **Stats** — live CPU, memory, network, and disk I/O (polled every 5s)
- **Prune** — clean up containers, images, volumes, and networks
- **Health endpoint** — `GET /healthz` for uptime monitors

## Stack

| Layer | Choice |
|-------|--------|
| Backend | Go + [Fiber v3](https://github.com/gofiber/fiber) |
| Docker | [Official SDK](https://pkg.go.dev/github.com/docker/docker/client) |
| Frontend | HTMX + Alpine.js + Tailwind CSS (CDN) |
| Templates | Go `html/template` via `go:embed` |

No JavaScript build step. No database. All state comes from the Docker daemon.

## Requirements

- Go 1.23+
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

- **Binary size:** ~14 MB
- **Idle RAM:** <20 MB
- **No streaming connections** — stats are polled snapshots, not persistent goroutines

## License

MIT
