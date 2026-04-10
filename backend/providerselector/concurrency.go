package providerselector

import (
	"fmt"
	"sync"
)

// InfraConcurrencyStats is a point-in-time snapshot of request tracking
// for a single infra provider.
type InfraConcurrencyStats struct {
	Name          string `json:"name"`
	Active        int    `json:"active"`
	TotalServed   int64  `json:"total_served"`
	MaxConcurrent int    `json:"max_concurrent"` // 0 = unlimited
	AtCapacity    bool   `json:"at_capacity"`
}

type infraCounter struct {
	mu            sync.Mutex
	active        int
	totalServed   int64
	maxConcurrent int // 0 = unlimited
}

// ConcurrencyManager tracks in-flight request counts for infra providers in
// memory. It is safe for concurrent use from multiple goroutines.
//
// Only infra providers participate in concurrency tracking; STT, LLM, and TTS
// providers are intentionally excluded.
type ConcurrencyManager struct {
	mu       sync.RWMutex
	counters map[string]*infraCounter
}

// NewConcurrencyManager creates a manager pre-seeded with the given provider
// names and their per-provider maximum concurrent request limits.
// A maxConcurrent value of 0 means the provider is treated as unlimited.
func NewConcurrencyManager(limits map[string]int) *ConcurrencyManager {
	counters := make(map[string]*infraCounter, len(limits))
	for name, max := range limits {
		counters[name] = &infraCounter{maxConcurrent: max}
	}
	return &ConcurrencyManager{counters: counters}
}

// RequestStart increments the active counter for name and the total-served
// lifetime counter. It returns an error if:
//   - the provider name is not registered, or
//   - the provider has a non-zero max and is already at capacity.
func (m *ConcurrencyManager) RequestStart(name string) error {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown infra provider: %q", name)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxConcurrent > 0 && c.active >= c.maxConcurrent {
		return fmt.Errorf("provider %q at capacity (%d/%d)", name, c.active, c.maxConcurrent)
	}
	c.active++
	c.totalServed++
	return nil
}

// RequestEnd decrements the active counter for name. It is safe to call even
// if the counter is already 0 (it will not go negative). Returns an error only
// when the provider name is not registered.
func (m *ConcurrencyManager) RequestEnd(name string) error {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown infra provider: %q", name)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active > 0 {
		c.active--
	}
	return nil
}

// Stats returns a snapshot of concurrency data for every registered infra
// provider. The order of entries is non-deterministic.
func (m *ConcurrencyManager) Stats() []InfraConcurrencyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]InfraConcurrencyStats, 0, len(m.counters))
	for name, c := range m.counters {
		c.mu.Lock()
		out = append(out, InfraConcurrencyStats{
			Name:          name,
			Active:        c.active,
			TotalServed:   c.totalServed,
			MaxConcurrent: c.maxConcurrent,
			AtCapacity:    c.maxConcurrent > 0 && c.active >= c.maxConcurrent,
		})
		c.mu.Unlock()
	}
	return out
}

// StatsFor returns the snapshot for a single provider. The second return value
// is false if the provider is not registered.
func (m *ConcurrencyManager) StatsFor(name string) (InfraConcurrencyStats, bool) {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return InfraConcurrencyStats{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return InfraConcurrencyStats{
		Name:          name,
		Active:        c.active,
		TotalServed:   c.totalServed,
		MaxConcurrent: c.maxConcurrent,
		AtCapacity:    c.maxConcurrent > 0 && c.active >= c.maxConcurrent,
	}, true
}

// DriftResult describes the difference between the in-memory active count and
// the actual run count reported by the real infrastructure.
type DriftResult struct {
	Name     string `json:"name"`
	InMemory int    `json:"in_memory"`
	Actual   int    `json:"actual"`
	Drift    int    `json:"drift"`     // Actual - InMemory; positive means more real runs than tracked
	HasDrift bool   `json:"has_drift"` // true when Drift != 0
}

// CheckDrift compares the in-memory active count for name against the provided
// actual count from the real infra. Returns an error if the provider is not registered.
func (m *ConcurrencyManager) CheckDrift(name string, actual int) (DriftResult, error) {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return DriftResult{}, fmt.Errorf("unknown infra provider: %q", name)
	}

	c.mu.Lock()
	inMemory := c.active
	c.mu.Unlock()

	drift := actual - inMemory
	return DriftResult{
		Name:     name,
		InMemory: inMemory,
		Actual:   actual,
		Drift:    drift,
		HasDrift: drift != 0,
	}, nil
}

// SyncActive forcefully sets the active counter for name to count.
// Used by the reconciler to correct drift between in-memory state and real infra.
// Returns an error if the provider is not registered.
func (m *ConcurrencyManager) SyncActive(name string, count int) error {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown infra provider: %q", name)
	}
	c.mu.Lock()
	c.active = count
	c.mu.Unlock()
	return nil
}

// IsAtCapacity returns true when the named provider has a non-zero max and
// its active count has reached that max. Unknown providers return false.
func (m *ConcurrencyManager) IsAtCapacity(name string) bool {
	m.mu.RLock()
	c, ok := m.counters[name]
	m.mu.RUnlock()
	if !ok {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxConcurrent > 0 && c.active >= c.maxConcurrent
}
