package checks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Runner struct {
	store          *Store
	client         *http.Client
	userAgent      string
	webhookURL     string
	webhookTimeout time.Duration

	mu      sync.Mutex
	nextRun map[string]time.Time
}

func NewRunner(store *Store, userAgent, webhookURL string, httpTimeoutSec, webhookTimeoutSec int) *Runner {
	httpTimeout := time.Duration(httpTimeoutSec) * time.Second
	if httpTimeout <= 0 {
		httpTimeout = 10 * time.Second
	}

	whTimeout := time.Duration(webhookTimeoutSec) * time.Second
	if whTimeout <= 0 {
		whTimeout = 5 * time.Second
	}

	return &Runner{
		store: store,
		client: &http.Client{
			Timeout: httpTimeout,
		},
		userAgent:      userAgent,
		webhookURL:     webhookURL,
		webhookTimeout: whTimeout,
		nextRun:        map[string]time.Time{},
	}
}

func (r *Runner) Start(ctx context.Context) {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.runDueChecks(ctx)
		}
	}
}

func (r *Runner) RunCheckNow(ctx context.Context, id string) (CheckResult, error) {
	check, err := r.store.Get(id)
	if err != nil {
		return CheckResult{}, err
	}

	result := r.execute(ctx, check)
	if err := r.store.SetLastResult(id, result); err != nil {
		return CheckResult{}, err
	}

	r.setNextRun(id, time.Now().UTC().Add(time.Duration(check.IntervalSec)*time.Second))
	r.maybeNotify(check, result)

	return result, nil
}

func (r *Runner) runDueChecks(ctx context.Context) {
	checks := r.store.List()
	now := time.Now().UTC()

	for _, check := range checks {
		if !check.Enabled {
			continue
		}
		if !r.isDue(check.ID, now) {
			continue
		}

		result := r.execute(ctx, check)
		_ = r.store.SetLastResult(check.ID, result)
		r.setNextRun(check.ID, now.Add(time.Duration(check.IntervalSec)*time.Second))
		r.maybeNotify(check, result)
	}
}

func (r *Runner) execute(ctx context.Context, check Check) CheckResult {
	start := time.Now()
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(check.TimeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, check.Method, check.URL, nil)
	if err != nil {
		return CheckResult{
			Status:    StatusDown,
			CheckedAt: time.Now().UTC(),
			Error:     err.Error(),
		}
	}

	req.Header.Set("User-Agent", r.userAgent)
	resp, err := r.client.Do(req)
	if err != nil {
		return CheckResult{
			Status:    StatusDown,
			CheckedAt: time.Now().UTC(),
			LatencyMS: time.Since(start).Milliseconds(),
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	status := StatusDown
	if resp.StatusCode == check.ExpectedStatus {
		status = StatusUp
	}

	return CheckResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		LatencyMS:  time.Since(start).Milliseconds(),
		CheckedAt:  time.Now().UTC(),
	}
}

func (r *Runner) isDue(id string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	next, ok := r.nextRun[id]
	if !ok {
		r.nextRun[id] = now
		return true
	}
	return !now.Before(next)
}

func (r *Runner) setNextRun(id string, t time.Time) {
	r.mu.Lock()
	r.nextRun[id] = t
	r.mu.Unlock()
}

func (r *Runner) maybeNotify(check Check, result CheckResult) {
	if r.webhookURL == "" {
		return
	}

	prev := StatusUnknown
	if check.LastResult != nil {
		prev = check.LastResult.Status
	}

	if prev == result.Status {
		return
	}

	if result.Status == StatusDown || (result.Status == StatusUp && check.NotifyOnRecovery) {
		_ = r.postWebhook(check, prev, result)
	}
}

func (r *Runner) postWebhook(check Check, prev Status, current CheckResult) error {
	payload := map[string]any{
		"checkId":        check.ID,
		"checkName":      check.Name,
		"url":            check.URL,
		"previousStatus": prev,
		"currentStatus":  current.Status,
		"statusCode":     current.StatusCode,
		"latencyMs":      current.LatencyMS,
		"checkedAt":      current.CheckedAt,
		"error":          current.Error,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.webhookTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.webhookURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}

	return nil
}
