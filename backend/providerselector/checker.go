package providerselector

import (
	"context"
	"log"
	"sync"
	"time"
)

type providerState struct {
	mu      sync.RWMutex
	results []CheckResult
	maxSize int
}

func (ps *providerState) add(r CheckResult) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.results) >= ps.maxSize {
		ps.results = ps.results[1:]
	}
	ps.results = append(ps.results, r)
}

func (ps *providerState) snapshot() []CheckResult {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make([]CheckResult, len(ps.results))
	copy(out, ps.results)
	return out
}

// Checker runs periodic health checks against each registered provider and
// maintains a rolling window of results.
type Checker struct {
	providers  []Provider
	states     map[string]*providerState
	interval   time.Duration
	windowSize int
}

// NewChecker creates a Checker from the given providers and config.
func NewChecker(providers []Provider, cfg *Config) *Checker {
	states := make(map[string]*providerState, len(providers))
	for _, p := range providers {
		key := stateKey(p)
		states[key] = &providerState{maxSize: cfg.Ranking.WindowSize}
	}
	return &Checker{
		providers:  providers,
		states:     states,
		interval:   cfg.CheckInterval,
		windowSize: cfg.Ranking.WindowSize,
	}
}

func stateKey(p Provider) string {
	return string(p.Type()) + "/" + p.Name()
}

// Start launches a background goroutine per provider. It returns immediately.
// The goroutines are cancelled when ctx is done.
func (c *Checker) Start(ctx context.Context) {
	for _, p := range c.providers {
		go c.runProvider(ctx, p)
	}
}

func (c *Checker) runProvider(ctx context.Context, p Provider) {
	state := c.states[stateKey(p)]

	doCheck := func() {
		log.Printf("[provider-health] checking %s/%s", p.Type(), p.Name())
		checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		r, err := p.Check(checkCtx)
		r.Timestamp = time.Now()
		if err != nil {
			r.Error = err.Error()
			r.Available = false
			log.Printf("[provider-health] %s/%s failed: %v", p.Type(), p.Name(), err)
		} else {
			log.Printf("[provider-health] %s/%s ok: available=%v latency=%s",
				p.Type(), p.Name(), r.Available, r.Latency.Round(time.Millisecond))
		}
		state.add(r)
	}

	doCheck() // immediate first check

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			doCheck()
		}
	}
}

// GetResults returns a snapshot of results for all providers of a given type.
func (c *Checker) GetResults(providerType ProviderType) map[string][]CheckResult {
	out := make(map[string][]CheckResult)
	for _, p := range c.providers {
		if p.Type() != providerType {
			continue
		}
		out[p.Name()] = c.states[stateKey(p)].snapshot()
	}
	return out
}

// GetAllResults returns all provider results grouped by type.
func (c *Checker) GetAllResults() map[ProviderType]map[string][]CheckResult {
	out := make(map[ProviderType]map[string][]CheckResult)
	for _, p := range c.providers {
		t := p.Type()
		if out[t] == nil {
			out[t] = make(map[string][]CheckResult)
		}
		out[t][p.Name()] = c.states[stateKey(p)].snapshot()
	}
	return out
}

// InjectResult adds an externally-observed result into the circular buffer.
// Returns false if the provider is not registered.
func (c *Checker) InjectResult(providerType ProviderType, name string, result CheckResult) bool {
	key := string(providerType) + "/" + name
	state, ok := c.states[key]
	if !ok {
		return false
	}
	if result.Timestamp.IsZero() {
		result.Timestamp = time.Now()
	}
	state.add(result)
	return true
}

// GetInfraOrder returns infra provider names in their registration order,
// which defines selection priority for the /best endpoint.
func (c *Checker) GetInfraOrder() []string {
	var names []string
	for _, p := range c.providers {
		if p.Type() == Infra {
			names = append(names, p.Name())
		}
	}
	return names
}

// GetAllMeta returns metadata from all registered providers, grouped by type.
func (c *Checker) GetAllMeta() map[ProviderType]map[string]map[string]string {
	out := make(map[ProviderType]map[string]map[string]string)
	for _, p := range c.providers {
		t := p.Type()
		if out[t] == nil {
			out[t] = make(map[string]map[string]string)
		}
		out[t][p.Name()] = p.Meta()
	}
	return out
}
