package systemd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// ServiceInfo holds the status of a systemd service.
type ServiceInfo struct {
	Name        string
	Description string
	ActiveState string // "active", "inactive", "failed"
	SubState    string // "running", "dead", "exited"
	MainPID     int
	MemCurrent  uint64 // bytes
	CPUNanos    uint64 // cumulative CPU nanoseconds
	StartedAt   string // human-readable timestamp
	NRestarts   int
	UnitEnabled string // "enabled", "disabled", "static"
}

// ValidName returns an error if the service name contains unsafe characters.
func ValidName(name string) error {
	if name == "" {
		return fmt.Errorf("empty service name")
	}
	if strings.ContainsAny(name, "/;`$|&\\ \t\n\"'") || strings.Contains(name, "..") {
		return fmt.Errorf("invalid service name: %q", name)
	}
	return nil
}

// ParseShow parses the key=value output of `systemctl show` into a ServiceInfo.
func ParseShow(name, output string) *ServiceInfo {
	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		k, v, ok := strings.Cut(line, "=")
		if ok {
			props[k] = v
		}
	}

	pid, _ := strconv.Atoi(props["MainPID"])
	mem := parseUint(props["MemoryCurrent"])
	cpu := parseUint(props["CPUUsageNSec"])
	restarts, _ := strconv.Atoi(props["NRestarts"])

	return &ServiceInfo{
		Name:        name,
		Description: props["Description"],
		ActiveState: props["ActiveState"],
		SubState:    props["SubState"],
		MainPID:     pid,
		MemCurrent:  mem,
		CPUNanos:    cpu,
		StartedAt:   props["ExecMainStartTimestamp"],
		NRestarts:   restarts,
		UnitEnabled: props["UnitFileState"],
	}
}

func parseUint(s string) uint64 {
	// MemoryCurrent and CPUUsageNSec may be "[not set]" if cgroup accounting is off
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
