package providerselector

import (
	"sort"
	"time"
)

// ProviderStats holds the computed health/ranking stats for one provider.
type ProviderStats struct {
	Name             string            `json:"name"`
	Score            float64           `json:"score"`
	AvgLatencyMs     float64           `json:"avg_latency_ms"`
	AvgTTFBMs        float64           `json:"avg_ttfb_ms,omitempty"`
	AvgTotalTimeMs   float64           `json:"avg_total_time_ms,omitempty"`
	Availability     float64           `json:"availability"`
	Checks           int               `json:"checks"`
	LastError        string            `json:"last_error,omitempty"`
	LastChecked      time.Time         `json:"last_checked,omitempty"`
	Meta             map[string]string `json:"meta,omitempty"`
	RecentChecks     []CheckResult     `json:"recent_checks,omitempty"`
	// Infra-only capacity fields (omitted for STT/LLM/TTS providers)
	ActiveRequests   *int `json:"active_requests,omitempty"`
	MaxConcurrent    *int `json:"max_concurrent,omitempty"`
	AvailableCapacity *int `json:"available_capacity,omitempty"`
}

// Ranker computes weighted scores for a set of providers.
type Ranker struct {
	latencyWeight      float64
	availabilityWeight float64
}

// NewRanker creates a Ranker from config.
func NewRanker(cfg *Config) *Ranker {
	return &Ranker{
		latencyWeight:      cfg.Ranking.LatencyWeight,
		availabilityWeight: cfg.Ranking.AvailabilityWeight,
	}
}

type rawStats struct {
	name         string
	avgLatency   float64
	avgTTFB      float64
	avgTotalTime float64
	availability float64
	checks       int
	lastError    string
	lastChecked  time.Time
	recentChecks []CheckResult
}

// Rank takes a map of name→[]CheckResult and returns sorted ProviderStats.
func (r *Ranker) Rank(results map[string][]CheckResult) []ProviderStats {
	raws := make([]rawStats, 0, len(results))
	minLatency := -1.0

	for name, checks := range results {
		if len(checks) == 0 {
			raws = append(raws, rawStats{name: name})
			continue
		}

		n := 5
		if len(checks) < n {
			n = len(checks)
		}
		recent := make([]CheckResult, n)
		for i := 0; i < n; i++ {
			recent[i] = checks[len(checks)-1-i]
		}

		var successCount int
		var totalLatency time.Duration
		var ttfbSum time.Duration
		var ttfbCount int
		var totalTimeSum time.Duration
		var lastError string
		lastChecked := checks[len(checks)-1].Timestamp

		for _, c := range checks {
			if c.Available {
				successCount++
				totalLatency += c.Latency
				if c.TTFB > 0 {
					ttfbSum += c.TTFB
					ttfbCount++
					totalTimeSum += c.TotalTime
				}
			}
			if c.Error != "" {
				lastError = c.Error
			}
		}

		availability := float64(successCount) / float64(len(checks))
		var avgLatency float64
		if successCount > 0 {
			avgLatency = float64(totalLatency/time.Millisecond) / float64(successCount)
		}

		if avgLatency > 0 && (minLatency < 0 || avgLatency < minLatency) {
			minLatency = avgLatency
		}

		var avgTTFB, avgTotalTime float64
		if ttfbCount > 0 {
			avgTTFB = float64(ttfbSum/time.Millisecond) / float64(ttfbCount)
			avgTotalTime = float64(totalTimeSum/time.Millisecond) / float64(ttfbCount)
		}

		raws = append(raws, rawStats{
			name:         name,
			avgLatency:   avgLatency,
			avgTTFB:      avgTTFB,
			avgTotalTime: avgTotalTime,
			availability: availability,
			checks:       len(checks),
			lastError:    lastError,
			lastChecked:  lastChecked,
			recentChecks: recent,
		})
	}

	stats := make([]ProviderStats, 0, len(raws))
	for _, raw := range raws {
		var normLatency float64
		if raw.avgLatency > 0 && minLatency > 0 {
			normLatency = minLatency / raw.avgLatency
		}
		score := r.latencyWeight*normLatency + r.availabilityWeight*raw.availability
		stats = append(stats, ProviderStats{
			Name:           raw.name,
			Score:          score,
			AvgLatencyMs:   raw.avgLatency,
			AvgTTFBMs:      raw.avgTTFB,
			AvgTotalTimeMs: raw.avgTotalTime,
			Availability:   raw.availability,
			Checks:         raw.checks,
			LastError:      raw.lastError,
			LastChecked:    raw.lastChecked,
			RecentChecks:   raw.recentChecks,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Score > stats[j].Score
	})

	return stats
}
