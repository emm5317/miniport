package handler

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
)

var (
	reTimestamp = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}\S*)`)
	reKV        = regexp.MustCompile(`(\w+)=("(?:[^"\\]|\\.)*"|\S+)`)
	hlTerms     = []string{"Session done", "error", "Error", "ERROR", "warn", "WARN", "failed", "FATAL", "fatal", "panic"}
)

// colorizeLogs colorizes all log lines and returns trusted HTML.
func colorizeLogs(raw string) template.HTML {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		lines[i] = colorizeLogLine(line)
	}
	return template.HTML(strings.Join(lines, "\n"))
}

// colorizeLogLine wraps timestamps and key=value pairs in colored spans.
// It HTML-escapes the line first, then applies trusted span tags.
func colorizeLogLine(line string) string {
	if line == "" {
		return line
	}
	// Escape HTML entities first to prevent XSS
	line = html.EscapeString(line)

	// Wrap leading timestamp
	line = reTimestamp.ReplaceAllString(line,
		`<span class="log-ts">$1</span>`)

	// Wrap key=value pairs
	line = reKV.ReplaceAllStringFunc(line, func(m string) string {
		parts := strings.SplitN(m, "=", 2)
		if len(parts) != 2 {
			return m
		}
		key, val := parts[0], parts[1]
		valClass := "log-val"
		for _, term := range hlTerms {
			if strings.Contains(val, term) {
				valClass = "log-hl"
				break
			}
		}
		return fmt.Sprintf(`<span class="log-key">%s</span>=<span class="%s">%s</span>`,
			key, valClass, val)
	})

	// Escalate full error lines
	if strings.Contains(line, "level=error") || strings.Contains(line, "level=ERROR") ||
		strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
		return `<span class="log-err-line">` + line + `</span>`
	}
	return line
}
