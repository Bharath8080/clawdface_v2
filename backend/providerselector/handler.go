package providerselector

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

// Handler holds references to the checker and ranker for serving HTTP requests.
type Handler struct {
	checker        *Checker
	ranker         *Ranker
	concurrency    *ConcurrencyManager
	activeCounters map[string]ActiveRunCounter
}

// NewHandler creates a Handler backed by the given Checker, Config, and
// ConcurrencyManager. Pass nil for concurrency to disable tracking.
// activeCounters maps provider names to their ActiveRunCounter implementations;
// pass nil to disable the infra reconciliation check endpoint.
func NewHandler(chk *Checker, cfg *Config, cm *ConcurrencyManager, activeCounters map[string]ActiveRunCounter) *Handler {
	return &Handler{
		checker:        chk,
		ranker:         NewRanker(cfg),
		concurrency:    cm,
		activeCounters: activeCounters,
	}
}

func (h *Handler) ranked() map[string][]ProviderStats {
	allResults := h.checker.GetAllResults()
	allMeta := h.checker.GetAllMeta()
	response := make(map[string][]ProviderStats)
	for _, t := range []ProviderType{STT, LLM, TTS, Infra} {
		stats := h.ranker.Rank(allResults[t])
		for i := range stats {
			stats[i].Meta = allMeta[t][stats[i].Name]
			if t == Infra && h.concurrency != nil {
				if cs, ok := h.concurrency.StatsFor(stats[i].Name); ok {
					active := cs.Active
					max := cs.MaxConcurrent
					stats[i].ActiveRequests = &active
					stats[i].MaxConcurrent = &max
					avail := max - active
					if max == 0 {
						avail = -1 // sentinel: unlimited
					}
					stats[i].AvailableCapacity = &avail
				}
			}
		}
		response[string(t)] = stats
	}
	return response
}

// HandleRanked returns all providers ranked by score, grouped by type.
// GET /api/v1/provider-health/ranked
func (h *Handler) HandleRanked(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.ranked())
}

type bestEntry struct {
	Name              string            `json:"name"`
	Meta              map[string]string `json:"meta,omitempty"`
	ActiveRequests    *int              `json:"active_requests,omitempty"`
	MaxConcurrent     *int              `json:"max_concurrent,omitempty"`
	AvailableCapacity *int              `json:"available_capacity,omitempty"`
}

// HandleBest returns the top-scoring available provider for each type.
// For infra providers, capacity is also considered: a provider at its max
// concurrent limit is skipped in favour of the next ranked candidate.
// GET /api/v1/provider-health/best
func (h *Handler) HandleBest(w http.ResponseWriter, r *http.Request) {
	ranked := h.ranked()
	best := make(map[string]*bestEntry)
	for _, t := range []ProviderType{STT, LLM, TTS} {
		for _, p := range ranked[string(t)] {
			if p.Availability > 0 {
				best[string(t)] = &bestEntry{Name: p.Name, Meta: p.Meta}
				break
			}
		}
	}
	// For infra, selection follows registration order (explicit priority) rather
	// than score-based ranking. Providers at their concurrency limit are skipped.
	infraStats := make(map[string]ProviderStats)
	for _, p := range ranked[string(Infra)] {
		infraStats[p.Name] = p
	}
	for _, name := range h.checker.GetInfraOrder() {
		p, ok := infraStats[name]
		if !ok || p.Availability <= 0 {
			continue
		}
		if h.concurrency != nil && h.concurrency.IsAtCapacity(p.Name) {
			continue
		}
		entry := &bestEntry{Name: p.Name, Meta: p.Meta}
		if h.concurrency != nil {
			if stats, ok := h.concurrency.StatsFor(p.Name); ok {
				active := stats.Active
				max := stats.MaxConcurrent
				entry.ActiveRequests = &active
				entry.MaxConcurrent = &max
				avail := max - active
				if max == 0 {
					avail = -1 // sentinel: unlimited
				}
				entry.AvailableCapacity = &avail
			}
		}
		best[string(Infra)] = entry
		break
	}
	writeJSON(w, http.StatusOK, best)
}

type statusEntry struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Results []CheckResult `json:"results"`
}

// HandleStatus returns all raw check results for every provider.
// GET /api/v1/provider-health/status
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	allResults := h.checker.GetAllResults()
	var entries []statusEntry
	for _, t := range []ProviderType{STT, LLM, TTS, Infra} {
		for name, results := range allResults[t] {
			entries = append(entries, statusEntry{
				Name:    name,
				Type:    string(t),
				Results: results,
			})
		}
	}
	writeJSON(w, http.StatusOK, entries)
}

type reportRequest struct {
	Type        ProviderType `json:"type"`
	Name        string       `json:"name"`
	LatencyMs   float64      `json:"latency_ms"`
	TTFBMs      float64      `json:"ttfb_ms,omitempty"`
	TotalTimeMs float64      `json:"total_time_ms,omitempty"`
	Available   bool         `json:"available"`
	Error       string       `json:"error,omitempty"`
}

// HandleReport accepts externally-observed check results for infra providers.
// POST /api/v1/provider-health/report
func (h *Handler) HandleReport(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if req.Type != Infra {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only infra type is accepted"})
		return
	}
	result := CheckResult{
		Latency:   time.Duration(req.LatencyMs * float64(time.Millisecond)),
		TTFB:      time.Duration(req.TTFBMs * float64(time.Millisecond)),
		TotalTime: time.Duration(req.TotalTimeMs * float64(time.Millisecond)),
		Available: req.Available,
		Error:     req.Error,
	}
	if !h.checker.InjectResult(req.Type, req.Name, result) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "provider not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type concurrencyActionRequest struct {
	Name string `json:"name"`
}

// HandleInfraRequestStart records that a new request has started on the named
// infra provider. Returns 429 if the provider is at its concurrency limit.
// POST /api/v1/provider-health/infra/request/start
func (h *Handler) HandleInfraRequestStart(w http.ResponseWriter, r *http.Request) {
	if h.concurrency == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "concurrency tracking not enabled"})
		return
	}
	var req concurrencyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if err := h.concurrency.RequestStart(req.Name); err != nil {
		// Distinguish capacity errors (provider known but full) from unknown provider.
		if stats, ok := h.concurrency.StatsFor(req.Name); ok && stats.AtCapacity {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	stats, _ := h.concurrency.StatsFor(req.Name)
	writeJSON(w, http.StatusOK, stats)
}

// HandleInfraRequestEnd records that a request has completed on the named
// infra provider.
// POST /api/v1/provider-health/infra/request/end
func (h *Handler) HandleInfraRequestEnd(w http.ResponseWriter, r *http.Request) {
	if h.concurrency == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "concurrency tracking not enabled"})
		return
	}
	var req concurrencyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if err := h.concurrency.RequestEnd(req.Name); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	stats, _ := h.concurrency.StatsFor(req.Name)
	writeJSON(w, http.StatusOK, stats)
}

// HandleInfraConcurrency returns live concurrency stats for all infra providers.
// GET /api/v1/provider-health/infra/concurrency
func (h *Handler) HandleInfraConcurrency(w http.ResponseWriter, r *http.Request) {
	if h.concurrency == nil {
		writeJSON(w, http.StatusOK, []InfraConcurrencyStats{})
		return
	}
	writeJSON(w, http.StatusOK, h.concurrency.Stats())
}

// HandleDashboard serves an HTML dashboard showing live provider health status.
// GET /api/v1/provider-health/dashboard
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	ranked := h.ranked()
	type providerRow struct {
		Name           string
		Score          float64
		Availability   float64
		AvgLatencyMs   float64
		Checks         int
		Meta           map[string]string
		IsBest         bool
		PriorityLabel  string
		ActiveRequests int
		MaxConcurrent  int
		IsInfra        bool
	}
	type categorySection struct {
		Label     string
		Type      string
		Providers []providerRow
	}
	var sections []categorySection
	labels := map[ProviderType]string{
		STT:   "Speech-to-Text",
		LLM:   "Language Models",
		TTS:   "Text-to-Speech",
		Infra: "Infrastructure",
	}
	priorityLabels := []string{"FIRST", "SECOND", "THIRD"}
	for _, t := range []ProviderType{STT, LLM, TTS, Infra} {
		var rows []providerRow

		if t == Infra {
			// Infra uses fixed priority order, not score ranking.
			infraStatsMap := make(map[string]ProviderStats)
			for _, p := range ranked[string(Infra)] {
				infraStatsMap[p.Name] = p
			}
			for i, name := range h.checker.GetInfraOrder() {
				p, ok := infraStatsMap[name]
				if !ok {
					continue
				}
				label := ""
				if i < len(priorityLabels) {
					label = priorityLabels[i]
				}
				row := providerRow{
					Name:          p.Name,
					Score:         p.Score,
					Availability:  p.Availability * 100,
					AvgLatencyMs:  p.AvgLatencyMs,
					Checks:        p.Checks,
					Meta:          p.Meta,
					PriorityLabel: label,
					IsInfra:       true,
				}
				if h.concurrency != nil {
					if stats, ok := h.concurrency.StatsFor(p.Name); ok {
						row.ActiveRequests = stats.Active
						row.MaxConcurrent = stats.MaxConcurrent
					}
				}
				rows = append(rows, row)
			}
		} else {
			for i, p := range ranked[string(t)] {
				rows = append(rows, providerRow{
					Name:         p.Name,
					Score:        p.Score,
					Availability: p.Availability * 100,
					AvgLatencyMs: p.AvgLatencyMs,
					Checks:       p.Checks,
					Meta:         p.Meta,
					IsBest:       i == 0 && p.Availability > 0,
				})
			}
		}

		sections = append(sections, categorySection{
			Label:     labels[t],
			Type:      string(t),
			Providers: rows,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	const tmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Provider Health Dashboard</title>
<style>
  *, *::before, *::after { box-sizing: border-box; }
  body { font-family: system-ui, -apple-system, sans-serif; margin: 0; padding: 2rem; background: #0f1117; color: #e2e8f0; min-height: 100vh; }
  h1 { font-size: 1.5rem; font-weight: 700; margin: 0 0 .25rem; color: #f8fafc; }
  .subtitle { font-size: .85rem; color: #64748b; margin-bottom: 2rem; }
  .refresh { font-size: .8rem; color: #475569; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 1.5rem; }
  .section { background: #1e2330; border: 1px solid #2d3748; border-radius: 12px; overflow: hidden; }
  .section-header { padding: .75rem 1.1rem; background: #161b27; border-bottom: 1px solid #2d3748; display: flex; align-items: center; gap: .5rem; }
  .section-label { font-weight: 700; font-size: .95rem; color: #94a3b8; text-transform: uppercase; letter-spacing: .05em; }
  .type-pill { font-size: .7rem; font-weight: 600; padding: .15em .5em; border-radius: 999px; background: #2d3748; color: #7c8ea6; margin-left: auto; }
  .card { padding: .85rem 1.1rem; border-bottom: 1px solid #1a1f2e; transition: background .15s; }
  .card:last-child { border-bottom: none; }
  .card:hover { background: #252b3b; }
  .card.best { background: #1a2540; border-left: 3px solid #3b82f6; }
  .card.best:hover { background: #1e2b4a; }
  .card-top { display: flex; align-items: center; gap: .5rem; margin-bottom: .4rem; }
  .provider-name { font-weight: 600; font-size: .95rem; color: #e2e8f0; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .best-badge { font-size: .65rem; font-weight: 700; padding: .15em .5em; border-radius: 4px; background: #1d4ed8; color: #bfdbfe; letter-spacing: .04em; }
  .priority-badge { font-size: .65rem; font-weight: 700; padding: .15em .5em; border-radius: 4px; background: #292524; color: #a8a29e; letter-spacing: .04em; }
  .status-badge { font-size: .7rem; font-weight: 700; padding: .2em .55em; border-radius: 4px; }
  .up   { background: #14532d; color: #86efac; }
  .down { background: #450a0a; color: #fca5a5; }
  .card-stats { display: flex; gap: 1.2rem; flex-wrap: wrap; }
  .stat { display: flex; flex-direction: column; }
  .stat-label { font-size: .68rem; color: #475569; text-transform: uppercase; letter-spacing: .05em; }
  .stat-value { font-size: .88rem; font-weight: 600; color: #cbd5e1; }
  .stat-value.good { color: #4ade80; }
  .stat-value.warn { color: #facc15; }
  .stat-value.bad  { color: #f87171; }
  .meta-row { margin-top: .45rem; display: flex; flex-wrap: wrap; gap: .25rem; }
  .tag { background: #1e293b; border: 1px solid #334155; color: #64748b; border-radius: 4px; padding: .1em .4em; font-size: .72em; }
  .empty { padding: 1.5rem 1.1rem; color: #475569; font-size: .85rem; }
</style>
<script>setTimeout(()=>location.reload(), 15000);</script>
</head>
<body>
<h1>Provider Health Dashboard</h1>
<p class="subtitle">Live status &amp; ranking &mdash; <span class="refresh">auto-refresh every 15s</span></p>
<div class="grid">
{{range .}}
  <div class="section">
    <div class="section-header">
      <span class="section-label">{{.Label}}</span>
      <span class="type-pill">{{.Type}}</span>
    </div>
    {{if .Providers}}
    {{range .Providers}}
    <div class="card{{if .IsBest}} best{{end}}">
      <div class="card-top">
        <span class="provider-name">{{.Name}}</span>
        {{if .PriorityLabel}}<span class="priority-badge">{{.PriorityLabel}}</span>
        {{else if .IsBest}}<span class="best-badge">BEST</span>{{end}}
        {{if gt .Availability 0.0}}<span class="status-badge up">UP</span>{{else}}<span class="status-badge down">DOWN</span>{{end}}
      </div>
      <div class="card-stats">
        {{if not .IsInfra}}<div class="stat">
          <span class="stat-label">Score</span>
          <span class="stat-value">{{printf "%.3f" .Score}}</span>
        </div>{{end}}
        <div class="stat">
          <span class="stat-label">Availability</span>
          <span class="stat-value{{if ge .Availability 90.0}} good{{else if ge .Availability 50.0}} warn{{else}} bad{{end}}">{{printf "%.1f" .Availability}}%</span>
        </div>
        <div class="stat">
          <span class="stat-label">Avg Latency</span>
          <span class="stat-value">{{printf "%.0f" .AvgLatencyMs}} ms</span>
        </div>
        <div class="stat">
          <span class="stat-label">Checks</span>
          <span class="stat-value">{{.Checks}}</span>
        </div>
        {{if .IsInfra}}<div class="stat">
          <span class="stat-label">Active Requests</span>
          <span class="stat-value">{{if gt .MaxConcurrent 0}}{{.ActiveRequests}}/{{.MaxConcurrent}}{{else}}{{.ActiveRequests}}{{end}}</span>
        </div>{{end}}
      </div>
      {{if .Meta}}<div class="meta-row">{{range $k,$v := .Meta}}<span class="tag">{{$k}}={{$v}}</span>{{end}}</div>{{end}}
    </div>
    {{end}}
    {{else}}<div class="empty">No providers registered.</div>{{end}}
  </div>
{{end}}
</div>
</body>
</html>`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, sections)
}

type driftCheckResult struct {
	DriftResult
	Error string `json:"error,omitempty"`
}

// HandleInfraConcurrencyCheck queries the real infra (e.g. Cerebrium) for each
// registered provider and compares the live run count to the in-memory counter,
// returning any drift between the two.
// GET /api/v1/provider-health/infra/concurrency/check
func (h *Handler) HandleInfraConcurrencyCheck(w http.ResponseWriter, r *http.Request) {
	if h.concurrency == nil || len(h.activeCounters) == 0 {
		writeJSON(w, http.StatusOK, []driftCheckResult{})
		return
	}

	var results []driftCheckResult
	for name, counter := range h.activeCounters {
		actual, err := counter.ActiveRunCount(r.Context())
		if err != nil {
			results = append(results, driftCheckResult{
				DriftResult: DriftResult{Name: name},
				Error:       err.Error(),
			})
			continue
		}
		dr, err := h.concurrency.CheckDrift(name, actual)
		if err != nil {
			results = append(results, driftCheckResult{
				DriftResult: DriftResult{Name: name},
				Error:       err.Error(),
			})
			continue
		}
		results = append(results, driftCheckResult{DriftResult: dr})
	}
	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
