package providerselector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ActiveRunCounter can query the real infrastructure to get the number of
// currently active (in-flight) runs. Implemented by infra providers that
// expose a run-listing API.
type ActiveRunCounter interface {
	ActiveRunCount(ctx context.Context) (int, error)
}

// --- Cerebrium ---

type infraCerebrium struct {
	name      string
	url       string
	projectID string
	appID     string
	apiToken  string
}

func newInfraCerebrium(name, url, projectID, appID, apiToken string) *infraCerebrium {
	return &infraCerebrium{
		name:      name,
		url:       url,
		projectID: projectID,
		appID:     appID,
		apiToken:  apiToken,
	}
}

func (c *infraCerebrium) Name() string            { return c.name }
func (c *infraCerebrium) Type() ProviderType      { return Infra }
func (c *infraCerebrium) Meta() map[string]string { return map[string]string{"url": c.url} }

func (c *infraCerebrium) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return CheckResult{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

// ActiveRunCount queries the Cerebrium REST API for the number of currently
// active runs (status RUNNING or PENDING). Returns an error if project_id,
// app_id, or api_token are not configured.
func (c *infraCerebrium) ActiveRunCount(ctx context.Context) (int, error) {
	if c.projectID == "" || c.appID == "" {
		return 0, fmt.Errorf("cerebrium provider %q: project_id and app_id must be configured for active-run checks", c.name)
	}
	url := fmt.Sprintf(
		"https://rest.cerebrium.ai/v2/projects/%s/apps/%s/runs",
		c.projectID, c.appID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("cerebrium runs API: HTTP %d", resp.StatusCode)
	}

	var body struct {
		Runs []struct {
			Status string `json:"status"`
		} `json:"runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("decoding cerebrium runs response: %w", err)
	}

	count := 0
	for _, run := range body.Runs {
		s := strings.ToUpper(run.Status)
		if s == "RUNNING" || s == "PENDING" {
			count++
		}
	}
	return count, nil
}

// --- Runpod ---

type infraRunpod struct {
	name     string
	url      string
	apiToken string
}

func newInfraRunpod(name, url, apiToken string) *infraRunpod {
	return &infraRunpod{name: name, url: url, apiToken: apiToken}
}

func (r *infraRunpod) Name() string            { return r.name }
func (r *infraRunpod) Type() ProviderType      { return Infra }
func (r *infraRunpod) Meta() map[string]string { return map[string]string{"url": r.url} }

// runpodHealthResponse mirrors the Runpod /health endpoint payload.
type runpodHealthResponse struct {
	Jobs struct {
		Completed int `json:"completed"`
		Failed    int `json:"failed"`
		InProgress int `json:"inProgress"`
		InQueue   int `json:"inQueue"`
		Retried   int `json:"retried"`
	} `json:"jobs"`
	Workers struct {
		Idle          int `json:"idle"`
		Initializing  int `json:"initializing"`
		Ready         int `json:"ready"`
		Running       int `json:"running"`
		Throttled     int `json:"throttled"`
		Unhealthy     int `json:"unhealthy"`
	} `json:"workers"`
}

// Check calls the Runpod health endpoint. The provider is considered available
// when the response is HTTP 200 and at least one worker is in the ready state.
func (r *infraRunpod) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return CheckResult{}, err
	}
	if r.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	latency := time.Since(start)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CheckResult{Latency: latency, Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var body runpodHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return CheckResult{Latency: latency, Available: false}, fmt.Errorf("decoding runpod health response: %w", err)
	}

	if body.Workers.Ready <= 0 {
		return CheckResult{Latency: latency, Available: false}, fmt.Errorf("no ready workers (ready=%d)", body.Workers.Ready)
	}
	return CheckResult{Latency: latency, Available: true}, nil
}

// ActiveRunCount returns the number of workers currently processing requests.
func (r *infraRunpod) ActiveRunCount(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return 0, err
	}
	if r.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("runpod health API: HTTP %d", resp.StatusCode)
	}

	var body runpodHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("decoding runpod health response: %w", err)
	}
	return body.Workers.Running, nil
}
