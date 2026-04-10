package providerselector

import (
	"context"
	"log"
	"time"
)

// StartReconciler starts a background goroutine that periodically queries the
// real infra for active run counts and reconciles the in-memory counters.
// When drift is detected it logs a warning and adjusts the counter to match.
// The goroutine stops when ctx is cancelled.
func StartReconciler(ctx context.Context, interval time.Duration, cm *ConcurrencyManager, counters map[string]ActiveRunCounter) {
	if cm == nil || len(counters) == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runReconcile(ctx, cm, counters)
			}
		}
	}()
}

func runReconcile(ctx context.Context, cm *ConcurrencyManager, counters map[string]ActiveRunCounter) {
	for name, counter := range counters {
		actual, err := counter.ActiveRunCount(ctx)
		if err != nil {
			log.Printf("[provider-health] reconciler: failed to get active run count for %q: %v", name, err)
			continue
		}
		dr, err := cm.CheckDrift(name, actual)
		if err != nil {
			log.Printf("[provider-health] reconciler: %v", err)
			continue
		}
		if !dr.HasDrift {
			continue
		}
		log.Printf("[provider-health] reconciler: drift on %q — in_memory=%d actual=%d drift=%+d, adjusting",
			name, dr.InMemory, dr.Actual, dr.Drift)
		if err := cm.SyncActive(name, actual); err != nil {
			log.Printf("[provider-health] reconciler: failed to sync %q: %v", name, err)
		}
	}
}
