package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryPool is a simple pre-warmed agent pool.
type InMemoryPool struct {
	factory AgentFactory
	warm    map[string][]Agent // spec key -> idle agents
	maxIdle int
	ttl     time.Duration
	mu      sync.Mutex
}

// AgentFactory creates a new agent from a spec.
type AgentFactory interface {
	Create(ctx context.Context, spec AgentSpec) (Agent, error)
}

// PoolOptions configures the pool.
type PoolOptions struct {
	MaxIdle int
	TTL     time.Duration
}

// NewPool creates a new agent pool.
func NewPool(factory AgentFactory, opts PoolOptions) *InMemoryPool {
	if opts.MaxIdle <= 0 {
		opts.MaxIdle = 4
	}
	if opts.TTL <= 0 {
		opts.TTL = 5 * time.Minute
	}
	return &InMemoryPool{
		factory: factory,
		warm:    make(map[string][]Agent),
		maxIdle: opts.MaxIdle,
		ttl:     opts.TTL,
	}
}

// Acquire gets an agent from the pool or creates a new one.
func (p *InMemoryPool) Acquire(ctx context.Context, spec AgentSpec) (Agent, error) {
	key := specKey(spec)

	p.mu.Lock()
	if len(p.warm[key]) > 0 {
		agent := p.warm[key][len(p.warm[key])-1]
		p.warm[key] = p.warm[key][:len(p.warm[key])-1]
		p.mu.Unlock()
		return agent, nil
	}
	p.mu.Unlock()

	return p.factory.Create(ctx, spec)
}

// Release returns an agent to the pool or closes it.
func (p *InMemoryPool) Release(ctx context.Context, agent Agent) error {
	key := specKeyFromAgent(agent)

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.warm[key]) >= p.maxIdle {
		return agent.Cancel(ctx)
	}

	p.warm[key] = append(p.warm[key], agent)

	agentID := agent.ID()
	time.AfterFunc(p.ttl, func() {
		p.evict(key, agentID)
	})

	return nil
}

// Warm pre-creates agents for a spec.
func (p *InMemoryPool) Warm(ctx context.Context, spec AgentSpec, count int) error {
	key := specKey(spec)
	agents := make([]Agent, 0, count)
	for i := 0; i < count; i++ {
		agent, err := p.factory.Create(ctx, spec)
		if err != nil {
			for _, a := range agents {
				_ = a.Cancel(ctx)
			}
			return err
		}
		agents = append(agents, agent)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.warm[key] = append(p.warm[key], agents...)
	return nil
}

// Stop closes all pooled agents.
func (p *InMemoryPool) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, agents := range p.warm {
		for _, a := range agents {
			_ = a.Cancel(ctx)
		}
	}
	p.warm = make(map[string][]Agent)
	return nil
}

func (p *InMemoryPool) evict(key, agentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, a := range p.warm[key] {
		if a.ID() == agentID {
			p.warm[key] = append(p.warm[key][:i], p.warm[key][i+1:]...)
			_ = a.Cancel(context.Background())
			break
		}
	}
}

func specKey(spec AgentSpec) string {
	return fmt.Sprintf("%s:%s:%s:%v", spec.Capability, spec.Sandbox, spec.Model, spec.Tools)
}

func specKeyFromAgent(a Agent) string {
	info := a.Info()
	if len(info.Capabilities) == 0 {
		return "default"
	}
	c := info.Capabilities[0]
	return fmt.Sprintf("%s:%s:%s:%v", c.Name, c.Sandbox, c.Model, c.Tools)
}
