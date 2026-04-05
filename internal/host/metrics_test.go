package host

import (
	"strings"
	"testing"
)

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		secs float64
		want string
	}{
		{0, "0m"},
		{300, "5m"},
		{3661, "1h 1m"},
		{90061, "1d 1h 1m"},
		{259200, "3d 0h 0m"},
	}
	for _, tt := range tests {
		got := FormatUptime(tt.secs)
		if got != tt.want {
			t.Errorf("FormatUptime(%v) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestParseCPULine(t *testing.T) {
	tests := []struct {
		line     string
		wantIdle uint64
		wantTot  uint64
	}{
		// user nice system idle iowait irq softirq steal
		{"cpu  10 5 3 80 2 0 0 0", 82, 100},
		{"cpu  100 0 50 800 50 0 0 0", 850, 1000},
	}
	for _, tt := range tests {
		idle, total, err := ParseCPULine(tt.line)
		if err != nil {
			t.Fatalf("ParseCPULine(%q) error: %v", tt.line, err)
		}
		if idle != tt.wantIdle {
			t.Errorf("idle = %d, want %d", idle, tt.wantIdle)
		}
		if total != tt.wantTot {
			t.Errorf("total = %d, want %d", total, tt.wantTot)
		}
	}
}

func TestParseCPULineErrors(t *testing.T) {
	_, _, err := ParseCPULine("cpu ")
	if err == nil {
		t.Error("expected error for too few fields")
	}
}

func TestParseMeminfo(t *testing.T) {
	content := `MemTotal:       16384000 kB
MemFree:         2000000 kB
MemAvailable:    8192000 kB
Buffers:          500000 kB
`
	snap := &HostSnapshot{}
	if err := ParseMeminfo(snap, strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}

	wantTotal := uint64(16384000) * 1024
	wantUsed := uint64(16384000-8192000) * 1024
	if snap.MemTotal != wantTotal {
		t.Errorf("MemTotal = %d, want %d", snap.MemTotal, wantTotal)
	}
	if snap.MemUsed != wantUsed {
		t.Errorf("MemUsed = %d, want %d", snap.MemUsed, wantUsed)
	}
	if snap.MemPercent < 49 || snap.MemPercent > 51 {
		t.Errorf("MemPercent = %.1f, want ~50", snap.MemPercent)
	}
}

func TestParseNetDev(t *testing.T) {
	content := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1000      10    0    0    0     0          0         0     1000      10    0    0    0     0       0          0
  eth0: 5000      50    0    0    0     0          0         0     3000      30    0    0    0     0       0          0
  eth1: 2000      20    0    0    0     0          0         0     1000      10    0    0    0     0       0          0
`
	rx, tx := ParseNetDev(strings.NewReader(content))
	if rx != 7000 {
		t.Errorf("NetRx = %d, want 7000", rx)
	}
	if tx != 4000 {
		t.Errorf("NetTx = %d, want 4000", tx)
	}
}
