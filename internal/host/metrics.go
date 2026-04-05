package host

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// HostSnapshot holds a point-in-time reading of host system metrics.
type HostSnapshot struct {
	CPUPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	MemPercent  float64
	DiskUsed    uint64
	DiskTotal   uint64
	DiskPercent float64
	NetRx       uint64
	NetTx       uint64
	Uptime      string // human-readable, e.g. "3d 12h 5m"
}

// FormatUptime converts seconds to a human-readable duration.
func FormatUptime(secs float64) string {
	total := int(secs)
	days := total / 86400
	hours := (total % 86400) / 3600
	mins := (total % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// ParseCPULine parses a "cpu ..." line into idle and total jiffies.
func ParseCPULine(line string) (idle, total uint64, err error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0, fmt.Errorf("too few fields: %d", len(fields))
	}
	var vals []uint64
	for _, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse %q: %w", f, err)
		}
		vals = append(vals, v)
	}
	for _, v := range vals {
		total += v
	}
	// idle is field index 3 (4th value after "cpu")
	idle = vals[3]
	if len(vals) > 4 {
		idle += vals[4] // iowait
	}
	return idle, total, nil
}

// ParseMeminfo reads MemTotal and MemAvailable from a reader.
func ParseMeminfo(snap *HostSnapshot, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var memTotal, memAvail uint64
	found := 0
	for scanner.Scan() && found < 2 {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			memTotal = parseKBLine(line)
			found++
		} else if strings.HasPrefix(line, "MemAvailable:") {
			memAvail = parseKBLine(line)
			found++
		}
	}
	snap.MemTotal = memTotal * 1024
	snap.MemUsed = (memTotal - memAvail) * 1024
	if memTotal > 0 {
		snap.MemPercent = float64(memTotal-memAvail) / float64(memTotal) * 100
	}
	return nil
}

func parseKBLine(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(fields[1], 10, 64)
	return v
}

// ParseNetDev sums rx/tx bytes across all non-loopback interfaces from a /proc/net/dev reader.
func ParseNetDev(r io.Reader) (rx, tx uint64) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue // skip headers
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64)
		t, _ := strconv.ParseUint(fields[8], 10, 64)
		rx += r
		tx += t
	}
	return rx, tx
}
