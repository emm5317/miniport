package handler

import (
	"strings"
	"testing"

	"github.com/emm5317/miniport/internal/docker"
)

func TestSparkline_Empty(t *testing.T) {
	got := string(Sparkline(nil))
	if !strings.Contains(got, "<svg") {
		t.Error("expected SVG element for nil history")
	}
	if strings.Contains(got, "polyline") {
		t.Error("expected no polyline for nil history")
	}
}

func TestSparkline_SinglePoint(t *testing.T) {
	got := string(Sparkline([]docker.StatsSnapshot{{CPUPercent: 50}}))
	if strings.Contains(got, "polyline") {
		t.Error("expected no polyline for single-point history")
	}
}

func TestSparkline_MultiplePoints(t *testing.T) {
	history := []docker.StatsSnapshot{
		{CPUPercent: 10},
		{CPUPercent: 50},
		{CPUPercent: 90},
	}
	got := string(Sparkline(history))
	if !strings.Contains(got, "polyline") {
		t.Error("expected polyline in sparkline SVG")
	}
	if !strings.Contains(got, "#34d399") {
		t.Error("expected green color for CPU sparkline")
	}
}

func TestSparklineMem_Color(t *testing.T) {
	history := []docker.StatsSnapshot{
		{MemPercent: 20},
		{MemPercent: 60},
	}
	got := string(SparklineMem(history))
	if !strings.Contains(got, "#22d3ee") {
		t.Error("expected cyan color for memory sparkline")
	}
}
