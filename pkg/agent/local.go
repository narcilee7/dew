package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

// LocalAgent is an in-process agent implementation.
type LocalAgent struct {
	id        string
	info      AgentInfo
	boundaries core.Boundaries
	session   session.Store
	tools     tools.ToolRegistry
}

// NewLocalAgent creates a new in-process agent.
func NewLocalAgent(id string, info AgentInfo, boundaries core.Boundaries, store session.Store, registry tools.ToolRegistry) *LocalAgent {
	return &LocalAgent{
		id:         id,
		info:       info,
		boundaries: boundaries,
		session:    store,
		tools:      registry,
	}
}

// ID returns the agent identifier.
func (a *LocalAgent) ID() string { return a.id }

// Info returns agent metadata.
func (a *LocalAgent) Info() AgentInfo { return a.info }

// Capabilities returns the agent capabilities.
func (a *LocalAgent) Capabilities() []Capability { return a.info.Capabilities }

// Run executes a task and returns the result.
func (a *LocalAgent) Run(ctx context.Context, task Task, events chan<- event.Event) (TaskResult, error) {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	sess, err := a.session.Create(ctx, session.CreateOptions{
		ID:       fmt.Sprintf("%s-sess-%d", a.id, time.Now().UnixNano()),
		ParentID: task.ParentID,
	})
	if err != nil {
		return TaskResult{TaskID: task.ID, Error: err.Error(), Status: core.StatusFailure}, err
	}
	defer sess.Close()

	// Seed context messages if provided.
	for _, m := range task.Context {
		if err := sess.Append(ctx, m); err != nil {
			return TaskResult{TaskID: task.ID, Error: err.Error(), Status: core.StatusFailure}, err
		}
	}

	// Seed user goal.
	if err := sess.Append(ctx, llm.Message{
		Role:    core.RoleUser,
		Content: task.Goal,
	}); err != nil {
		return TaskResult{TaskID: task.ID, Error: err.Error(), Status: core.StatusFailure}, err
	}

	opts := core.RunOptions{
		MaxTurns: task.MaxTurns,
		Timeout:  task.Timeout,
	}
	if opts.MaxTurns == 0 {
		opts.MaxTurns = 50
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}

	harness := core.NewHarness(a.id, a.boundaries)
	if events != nil {
		go func() {
			for ev := range harness.Events() {
				select {
				case events <- ev:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	err = harness.Run(ctx, sess, opts)

	// Build summary from last assistant message.
	var summary string
	for i := len(sess.Messages()) - 1; i >= 0; i-- {
		m := sess.Messages()[i]
		if m.Role == core.RoleAssistant {
			summary = m.Content
			break
		}
	}

	status := core.StatusSuccess
	if err != nil {
		status = core.StatusFailure
		summary = err.Error()
	}

	return TaskResult{
		TaskID:  task.ID,
		Status:  status,
		Summary: summary,
	}, err
}

// Spawn starts a task in the background and returns a handle.
func (a *LocalAgent) Spawn(ctx context.Context, task Task) (Handle, error) {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	h := &localHandle{
		id:     task.ID,
		done:   make(chan struct{}),
		result: make(chan TaskResult, 1),
	}

	go func() {
		events := make(chan event.Event, 64)
		go func() {
			for range events {
				// Background tasks currently discard events.
			}
		}()

		res, err := a.Run(ctx, task, events)
		if err != nil {
			res.Error = err.Error()
		}
		h.result <- res
		close(h.done)
	}()

	return h, nil
}

// Cancel stops the agent.
func (a *LocalAgent) Cancel(ctx context.Context) error {
	// Local agent cancellation is handled via the parent context.
	return nil
}

// localHandle is a reference to a running LocalAgent task.
type localHandle struct {
	id     string
	done   chan struct{}
	result chan TaskResult
	mu     sync.Mutex
	res    *TaskResult
}

// ID returns the task identifier.
func (h *localHandle) ID() string { return h.id }

// Wait blocks until the task completes and returns the result.
func (h *localHandle) Wait(ctx context.Context) (TaskResult, error) {
	select {
	case <-ctx.Done():
		return TaskResult{}, ctx.Err()
	case <-h.done:
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.res == nil {
			res := <-h.result
			h.res = &res
		}
		return *h.res, nil
	}
}

// Cancel requests task cancellation.
func (h *localHandle) Cancel(ctx context.Context) error {
	// Cancellation is handled via the parent context passed to Spawn.
	return nil
}
