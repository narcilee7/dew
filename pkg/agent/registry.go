package agent

import (
	"context"
	"fmt"
	"sync"
)

// LocalRegistry is an in-memory optional registry.
type LocalRegistry struct {
	agents map[string]AgentInfo
	mu     sync.RWMutex
}

// NewLocalRegistry creates a new in-memory registry.
func NewLocalRegistry() *LocalRegistry {
	return &LocalRegistry{
		agents: make(map[string]AgentInfo),
	}
}

// Register registers an agent.
func (r *LocalRegistry) Register(ctx context.Context, info AgentInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[info.ID] = info
	return nil
}

// Deregister removes an agent.
func (r *LocalRegistry) Deregister(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, id)
	return nil
}

// List returns agents matching the filter.
func (r *LocalRegistry) List(ctx context.Context, filter CapabilityFilter) ([]AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []AgentInfo
	for _, info := range r.agents {
		if filter.Name != "" {
			matched := false
			for _, c := range info.Capabilities {
				if c.Name == filter.Name {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(filter.Tools) > 0 {
			hasTool := false
			for _, c := range info.Capabilities {
				for _, t := range c.Tools {
					for _, ft := range filter.Tools {
						if t == ft {
							hasTool = true
							break
						}
					}
				}
			}
			if !hasTool {
				continue
			}
		}
		out = append(out, info)
	}
	return out, nil
}

// Resolve returns a specific agent by ID.
func (r *LocalRegistry) Resolve(ctx context.Context, id string) (AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.agents[id]
	if !ok {
		return AgentInfo{}, fmt.Errorf("agent %q not found", id)
	}
	return info, nil
}
