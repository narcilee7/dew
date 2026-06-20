package tui

import (
	"strings"
)

// View renders the TUI.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString("dew\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	// Show the most recent lines that fit in the viewport.
	visible := m.height - 4
	if visible < 1 {
		visible = 1
	}
	start := 0
	if len(m.lines) > visible {
		start = len(m.lines) - visible
	}
	for _, line := range m.lines[start:] {
		b.WriteString(truncate(line, m.width))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	if m.running {
		b.WriteString(spinnerFrames[m.spinner] + " running...")
	} else if m.err != nil {
		b.WriteString("finished with error")
	} else {
		b.WriteString("done")
	}
	return b.String()
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	// Simple byte-based truncation is sufficient for ASCII logs.
	if len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}
