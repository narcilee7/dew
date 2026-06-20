package core

// Boundary is the base abstraction for all isolatable resources in dew.
// Every resource that can be scoped to a session, sandbox, or agent implements Boundary.
type Boundary interface {
	// ID returns a stable identifier for this boundary instance.
	ID() string

	// Close releases any resources held by the boundary.
	Close() error
}
