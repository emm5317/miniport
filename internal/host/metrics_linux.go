//go:build linux

package host

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Snapshot reads current host metrics from /proc and syscall.
func Snapshot() (*HostSnapshot, error) {
	snap := &HostSnapshot{}

	cpu, err := readCPU()
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	snap.CPUPercent = cpu

	if err := readMeminfo(snap); err != nil {
		return nil, fmt.Errorf("meminfo: %w", err)
	}

	if err := readDisk(snap); err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}

	if err := readNet(snap); err != nil {
		return nil, fmt.Errorf("net: %w", err)
	}

	if err := readUptime(snap); err != nil {
		return nil, fmt.Errorf("uptime: %w", err)
	}

	return snap, nil
}

func readCPU() (float64, error) {
	idle1, total1, err := parseProcStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(100 * time.Millisecond)
	idle2, total2, err := parseProcStat()
	if err != nil {
		return 0, err
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	if totalDelta == 0 {
		return 0, nil
	}
	return (1 - idleDelta/totalDelta) * 100, nil
}

func parseProcStat() (idle, total uint64, err error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		return ParseCPULine(line)
	}
	return 0, 0, fmt.Errorf("cpu line not found")
}

func readMeminfo(snap *HostSnapshot) error {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return err
	}
	defer f.Close()
	return ParseMeminfo(snap, f)
}

func readDisk(snap *HostSnapshot) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return err
	}
	snap.DiskTotal = stat.Blocks * uint64(stat.Bsize)
	snap.DiskUsed = snap.DiskTotal - (stat.Bavail * uint64(stat.Bsize))
	if snap.DiskTotal > 0 {
		snap.DiskPercent = float64(snap.DiskUsed) / float64(snap.DiskTotal) * 100
	}
	return nil
}

func readNet(snap *HostSnapshot) error {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return err
	}
	defer f.Close()

	rx, tx := ParseNetDev(f)
	snap.NetRx = rx
	snap.NetTx = tx
	return nil
}

func readUptime(snap *HostSnapshot) error {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return fmt.Errorf("empty uptime")
	}
	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return err
	}
	snap.Uptime = FormatUptime(secs)
	return nil
}
