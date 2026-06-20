package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case eventMsg:
		m.lines = append(m.lines, formatEvent(msg.ev))
		m.trimLines()
		return m, readEvent(m.evCh)

	case doneMsg:
		m.running = false
		m.done = true
		m.err = msg.err
		if msg.err != nil {
			m.lines = append(m.lines, "error: "+msg.err.Error())
		}
		return m, nil

	case tickMsg:
		if m.running {
			m.spinner = (m.spinner + 1) % len(spinnerFrames)
			return m, tick()
		}
	}

	return m, nil
}

func (m *Model) trimLines() {
	max := 1024
	if len(m.lines) > max {
		m.lines = m.lines[len(m.lines)-max:]
	}
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
