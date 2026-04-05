package docker

import (
	"testing"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func TestCalculateStats_Normal(t *testing.T) {
	resp := &container.StatsResponse{}
	resp.CPUStats.CPUUsage.TotalUsage = 500
	resp.CPUStats.CPUUsage.PercpuUsage = make([]uint64, 4)
	resp.CPUStats.SystemUsage = 2000
	resp.PreCPUStats.CPUUsage.TotalUsage = 400
	resp.PreCPUStats.SystemUsage = 1000

	resp.MemoryStats.Usage = 1048576 // 1 MB
	resp.MemoryStats.Limit = 4194304 // 4 MB
	resp.MemoryStats.Stats = map[string]uint64{"cache": 0}

	resp.Networks = map[string]container.NetworkStats{
		"eth0": {RxBytes: 100, TxBytes: 200},
		"eth1": {RxBytes: 50, TxBytes: 75},
	}
	resp.BlkioStats.IoServiceBytesRecursive = []container.BlkioStatEntry{
		{Op: "Read", Value: 1024},
		{Op: "Write", Value: 2048},
		{Op: "read", Value: 512},
	}

	snap := calculateStats(resp)

	// CPU: (100/1000) * 4 * 100 = 40%
	if snap.CPUPercent < 39.9 || snap.CPUPercent > 40.1 {
		t.Errorf("CPUPercent = %f, want ~40.0", snap.CPUPercent)
	}
	if snap.MemUsage != 1048576 {
		t.Errorf("MemUsage = %d, want 1048576", snap.MemUsage)
	}
	if snap.MemLimit != 4194304 {
		t.Errorf("MemLimit = %d, want 4194304", snap.MemLimit)
	}
	if snap.MemPercent < 24.9 || snap.MemPercent > 25.1 {
		t.Errorf("MemPercent = %f, want ~25.0", snap.MemPercent)
	}
	if snap.NetRx != 150 {
		t.Errorf("NetRx = %d, want 150", snap.NetRx)
	}
	if snap.NetTx != 275 {
		t.Errorf("NetTx = %d, want 275", snap.NetTx)
	}
	if snap.BlockRead != 1536 {
		t.Errorf("BlockRead = %d, want 1536", snap.BlockRead)
	}
	if snap.BlockWrite != 2048 {
		t.Errorf("BlockWrite = %d, want 2048", snap.BlockWrite)
	}
}

func TestCalculateStats_ZeroDelta(t *testing.T) {
	resp := &container.StatsResponse{}
	// Same CPU values => sysDelta = 0 => cpuPercent = 0
	resp.CPUStats.CPUUsage.TotalUsage = 100
	resp.CPUStats.SystemUsage = 500
	resp.PreCPUStats.CPUUsage.TotalUsage = 100
	resp.PreCPUStats.SystemUsage = 500
	resp.MemoryStats.Stats = map[string]uint64{}

	snap := calculateStats(resp)
	if snap.CPUPercent != 0 {
		t.Errorf("CPUPercent = %f, want 0", snap.CPUPercent)
	}
}

func TestCalculateStats_CacheSubtracted(t *testing.T) {
	resp := &container.StatsResponse{}
	resp.MemoryStats.Usage = 2000
	resp.MemoryStats.Limit = 8000
	resp.MemoryStats.Stats = map[string]uint64{"cache": 500}

	snap := calculateStats(resp)
	if snap.MemUsage != 1500 {
		t.Errorf("MemUsage = %d, want 1500 (2000 - 500 cache)", snap.MemUsage)
	}
}

func TestCalculateStats_NoCacheKey(t *testing.T) {
	resp := &container.StatsResponse{}
	resp.MemoryStats.Usage = 2000
	resp.MemoryStats.Limit = 8000
	resp.MemoryStats.Stats = map[string]uint64{}

	snap := calculateStats(resp)
	if snap.MemUsage != 2000 {
		t.Errorf("MemUsage = %d, want 2000 (no cache key)", snap.MemUsage)
	}
}

func TestCalculateStats_ZeroMemLimit(t *testing.T) {
	resp := &container.StatsResponse{}
	resp.MemoryStats.Usage = 1000
	resp.MemoryStats.Limit = 0
	resp.MemoryStats.Stats = map[string]uint64{}

	snap := calculateStats(resp)
	if snap.MemPercent != 0 {
		t.Errorf("MemPercent = %f, want 0 (zero limit)", snap.MemPercent)
	}
}

func TestFormatPorts_Empty(t *testing.T) {
	result := formatPorts(nil)
	if result != "" {
		t.Errorf("formatPorts(nil) = %q, want empty", result)
	}
}

func TestFormatPorts_Single(t *testing.T) {
	ports := []dtypes.Port{{PublicPort: 8080, PrivatePort: 80}}
	result := formatPorts(ports)
	if result != "8080→80" {
		t.Errorf("formatPorts = %q, want %q", result, "8080→80")
	}
}

func TestFormatPorts_NoPublicPort(t *testing.T) {
	ports := []dtypes.Port{{PublicPort: 0, PrivatePort: 80}}
	result := formatPorts(ports)
	if result != "" {
		t.Errorf("formatPorts with no public port = %q, want empty", result)
	}
}

func TestSummarize(t *testing.T) {
	containers := []ContainerInfo{
		{State: "running", Status: "Up 2 hours"},
		{State: "running", Status: "Up 1 hour (unhealthy)"},
		{State: "exited", Status: "Exited (0) 3 hours ago"},
		{State: "dead", Status: "Dead"},
		{State: "restarting", Status: "Restarting"},
	}
	s := Summarize(containers)
	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.Running != 2 {
		t.Errorf("Running = %d, want 2", s.Running)
	}
	if s.Stopped != 2 {
		t.Errorf("Stopped = %d, want 2", s.Stopped)
	}
	if s.Unhealthy != 1 {
		t.Errorf("Unhealthy = %d, want 1", s.Unhealthy)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 || s.Running != 0 || s.Stopped != 0 || s.Unhealthy != 0 {
		t.Errorf("Expected all zeros for empty list, got %+v", s)
	}
}

func TestFormatPorts_Deduplication(t *testing.T) {
	ports := []dtypes.Port{
		{PublicPort: 8080, PrivatePort: 80},
		{PublicPort: 8080, PrivatePort: 80},
		{PublicPort: 3306, PrivatePort: 3306},
	}
	result := formatPorts(ports)
	if result != "8080→80, 3306→3306" {
		t.Errorf("formatPorts = %q, want %q", result, "8080→80, 3306→3306")
	}
}
