package checks

import "time"

const (
	MethodGet  = "GET"
	MethodHead = "HEAD"
)

type Status string

const (
	StatusUnknown Status = "unknown"
	StatusUp      Status = "up"
	StatusDown    Status = "down"
)

type Check struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	URL              string       `json:"url"`
	Method           string       `json:"method"`
	IntervalSec      int          `json:"intervalSec"`
	TimeoutSec       int          `json:"timeoutSec"`
	ExpectedStatus   int          `json:"expectedStatus"`
	NotifyOnRecovery bool         `json:"notifyOnRecovery"`
	Enabled          bool         `json:"enabled"`
	LastResult       *CheckResult `json:"lastResult,omitempty"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

type CheckResult struct {
	Status     Status    `json:"status"`
	StatusCode int       `json:"statusCode"`
	LatencyMS  int64     `json:"latencyMs"`
	CheckedAt  time.Time `json:"checkedAt"`
	Error      string    `json:"error,omitempty"`
}

type CreateCheckInput struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	Method           string `json:"method"`
	IntervalSec      int    `json:"intervalSec"`
	TimeoutSec       int    `json:"timeoutSec"`
	ExpectedStatus   int    `json:"expectedStatus"`
	NotifyOnRecovery bool   `json:"notifyOnRecovery"`
	Enabled          *bool  `json:"enabled,omitempty"`
}

type UpdateCheckInput struct {
	Name             *string `json:"name,omitempty"`
	URL              *string `json:"url,omitempty"`
	Method           *string `json:"method,omitempty"`
	IntervalSec      *int    `json:"intervalSec,omitempty"`
	TimeoutSec       *int    `json:"timeoutSec,omitempty"`
	ExpectedStatus   *int    `json:"expectedStatus,omitempty"`
	NotifyOnRecovery *bool   `json:"notifyOnRecovery,omitempty"`
	Enabled          *bool   `json:"enabled,omitempty"`
}
