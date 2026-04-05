//go:build linux

package systemd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	fastTimeout = 5 * time.Second
	slowTimeout = 30 * time.Second
)

var showProps = []string{
	"ActiveState", "SubState", "MainPID", "MemoryCurrent",
	"CPUUsageNSec", "ExecMainStartTimestamp", "NRestarts",
	"UnitFileState", "Description",
}

// Show returns the status of a systemd service.
func Show(ctx context.Context, name string) (*ServiceInfo, error) {
	if err := ValidName(name); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, fastTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "show", name, "--no-pager",
		"--property="+joinProps())
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("systemctl show %s: %w", name, err)
	}
	return ParseShow(name, out.String()), nil
}

func joinProps() string {
	return strings.Join(showProps, ",")
}

// Logs returns the last N lines of journal output for a service.
func Logs(ctx context.Context, name string, lines int) (string, error) {
	if err := ValidName(name); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "journalctl", "-u", name, "--no-pager",
		"-n", fmt.Sprintf("%d", lines), "-o", "short-iso")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("journalctl %s: %w", name, err)
	}
	return out.String(), nil
}

// LogsSince returns journal lines since the given timestamp.
func LogsSince(ctx context.Context, name, since string) (string, error) {
	if err := ValidName(name); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "journalctl", "-u", name, "--no-pager",
		"--since", since, "-o", "short-iso")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("journalctl %s: %w", name, err)
	}
	return out.String(), nil
}

// Start starts a systemd service.
func Start(ctx context.Context, name string) error {
	return systemctl(ctx, "start", name)
}

// Stop stops a systemd service.
func Stop(ctx context.Context, name string) error {
	return systemctl(ctx, "stop", name)
}

// Restart restarts a systemd service.
func Restart(ctx context.Context, name string) error {
	return systemctl(ctx, "restart", name)
}

func systemctl(ctx context.Context, action, name string) error {
	if err := ValidName(name); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, slowTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", action, name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, name, err)
	}
	return nil
}
