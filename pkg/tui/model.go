package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/session"
)

// Model is a Bubble Tea model that renders an agent run.
type Model struct {
	harness core.Harness
	sess    session.Session
	opts    core.RunOptions

	evCh   chan event.Event
	doneCh chan error

	width   int
	height  int
	lines   []string
	running bool
	done    bool
	err     error
	spinner int
}

// NewModel creates a TUI model for the given harness run.
func NewModel(harness core.Harness, sess session.Session, opts core.RunOptions) *Model {
	return &Model{
		harness: harness,
		sess:    sess,
		opts:    opts,
		running: true,
	}
}

// Init starts the harness run and event consumer.
func (m *Model) Init() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.evCh = make(chan event.Event, 256)
	m.doneCh = make(chan error, 1)

	go func() {
		m.doneCh <- m.harness.Run(ctx, m.sess, m.opts)
		cancel()
	}()

	go func() {
		for ev := range m.harness.Events() {
			m.evCh <- ev
		}
		close(m.evCh)
	}()

	return tea.Batch(
		readEvent(m.evCh),
		tick(),
		waitDone(m.doneCh),
	)
}

// eventMsg carries a single lifecycle event into the TUI.
type eventMsg struct {
	ev event.Event
}

// doneMsg signals that the harness run finished.
type doneMsg struct {
	err error
}

// tickMsg triggers a spinner frame update.
type tickMsg time.Time

func readEvent(ch chan event.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return eventMsg{ev: ev}
	}
}

func waitDone(ch chan error) tea.Cmd {
	return func() tea.Msg {
		return doneMsg{err: <-ch}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Run starts the TUI and blocks until the agent run completes.
func Run(harness core.Harness, sess session.Session, opts core.RunOptions) error {
	m := NewModel(harness, sess, opts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
