package handler

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/emm5317/miniport/internal/docker"
)

// Sparkline renders an SVG sparkline of CPU% from a stats history slice.
func Sparkline(history []docker.StatsSnapshot) template.HTML {
	return renderSparkline(history, func(s docker.StatsSnapshot) float64 { return s.CPUPercent }, "#34d399")
}

// SparklineMem renders an SVG sparkline of Mem% from a stats history slice.
func SparklineMem(history []docker.StatsSnapshot) template.HTML {
	return renderSparkline(history, func(s docker.StatsSnapshot) float64 { return s.MemPercent }, "#22d3ee")
}

func renderSparkline(history []docker.StatsSnapshot, getValue func(docker.StatsSnapshot) float64, color string) template.HTML {
	if len(history) < 2 {
		return template.HTML(`<svg viewBox="0 0 60 20" class="sparkline"></svg>`)
	}

	w, h := 60.0, 20.0
	points := make([]float64, len(history))
	maxVal := 100.0 // percentages are 0-100
	for i, s := range history {
		v := getValue(s)
		if v > maxVal {
			maxVal = v
		}
		points[i] = v
	}

	coords := make([]string, len(points))
	for i, v := range points {
		x := (float64(i) / float64(len(points)-1)) * w
		y := h - (v/maxVal)*h
		if y < 0.5 {
			y = 0.5
		}
		coords[i] = fmt.Sprintf("%.1f,%.1f", x, y)
	}

	path := strings.Join(coords, " ")
	svg := fmt.Sprintf(
		`<svg viewBox="0 0 60 20" class="sparkline"><polyline points="%s" fill="none" stroke="%s" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
		path, color,
	)
	return template.HTML(svg)
}
