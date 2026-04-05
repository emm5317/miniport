package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
)

// ContainerInfo is the view model for a container row.
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	State   string
	Status  string
	Ports   string
	Created int64
}

// Summary holds aggregate container counts for the dashboard strip.
type Summary struct {
	Total     int
	Running   int
	Stopped   int
	Unhealthy int
}

// Summarize computes aggregate counts from a container list.
func Summarize(containers []ContainerInfo) Summary {
	var s Summary
	s.Total = len(containers)
	for _, c := range containers {
		switch c.State {
		case "running":
			s.Running++
		case "exited", "dead":
			s.Stopped++
		}
		if strings.Contains(c.Status, "unhealthy") {
			s.Unhealthy++
		}
	}
	return s
}

// StatsSnapshot holds a single point-in-time resource reading.
type StatsSnapshot struct {
	CPUPercent float64
	MemUsage   uint64
	MemLimit   uint64
	MemPercent float64
	NetRx      uint64
	NetTx      uint64
	BlockRead  uint64
	BlockWrite uint64
}

const (
	defaultTimeout = 10 * time.Second
	slowTimeout    = 30 * time.Second
)

func (s *Service) List(ctx context.Context) ([]ContainerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	raw, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	out := make([]ContainerInfo, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		out = append(out, ContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Ports:   formatPorts(c.Ports),
			Created: c.Created,
		})
	}
	return out, nil
}

func (s *Service) Start(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	return s.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (s *Service) Stop(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()
	return s.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (s *Service) Restart(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()
	return s.cli.ContainerRestart(ctx, id, container.StopOptions{})
}

func (s *Service) Remove(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	return s.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (s *Service) Logs(ctx context.Context, id string, lines int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()

	reader, err := s.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", lines),
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return "", err
	}

	out := stdout.String()
	if stderr.Len() > 0 {
		out += "\n--- stderr ---\n" + stderr.String()
	}
	return out, nil
}

func (s *Service) Stats(ctx context.Context, id string) (*StatsSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	resp, err := s.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}
	return calculateStats(&stats), nil
}

func (s *Service) PruneContainers(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	report, err := s.cli.ContainersPrune(ctx, filters.NewArgs())
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

func (s *Service) PruneImages(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()
	report, err := s.cli.ImagesPrune(ctx, filters.NewArgs())
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

func (s *Service) PruneVolumes(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	report, err := s.cli.VolumesPrune(ctx, filters.NewArgs())
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

func (s *Service) PruneNetworks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	_, err := s.cli.NetworksPrune(ctx, filters.NewArgs())
	return err
}

func calculateStats(s *container.StatsResponse) *StatsSnapshot {
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage - s.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if sysDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / sysDelta) * float64(len(s.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	memUsage := s.MemoryStats.Usage
	if cache, ok := s.MemoryStats.Stats["cache"]; ok {
		memUsage -= cache
	}
	memPercent := 0.0
	if s.MemoryStats.Limit > 0 {
		memPercent = float64(memUsage) / float64(s.MemoryStats.Limit) * 100.0
	}

	var netRx, netTx uint64
	for _, v := range s.Networks {
		netRx += v.RxBytes
		netTx += v.TxBytes
	}

	var blockRead, blockWrite uint64
	for _, bio := range s.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(bio.Op) {
		case "read":
			blockRead += bio.Value
		case "write":
			blockWrite += bio.Value
		}
	}

	return &StatsSnapshot{
		CPUPercent: cpuPercent,
		MemUsage:   memUsage,
		MemLimit:   s.MemoryStats.Limit,
		MemPercent: memPercent,
		NetRx:      netRx,
		NetTx:      netTx,
		BlockRead:  blockRead,
		BlockWrite: blockWrite,
	}
}

func formatPorts(ports []dtypes.Port) string {
	seen := make(map[string]bool)
	var parts []string
	for _, p := range ports {
		if p.PublicPort == 0 {
			continue
		}
		s := fmt.Sprintf("%d→%d", p.PublicPort, p.PrivatePort)
		if !seen[s] {
			seen[s] = true
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ", ")
}
