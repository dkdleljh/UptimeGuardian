package checks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("check not found")

type Store struct {
	mu       sync.RWMutex
	checks   map[string]Check
	dataFile string
}

func NewStore(dataFile string) (*Store, error) {
	s := &Store{
		checks:   make(map[string]Check),
		dataFile: dataFile,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) List() []Check {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Check, 0, len(s.checks))
	for _, c := range s.checks {
		items = append(items, c)
	}

	// Stable ordering for clients.
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items
}

func (s *Store) Get(id string) (Check, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.checks[id]
	if !ok {
		return Check{}, ErrNotFound
	}
	return c, nil
}

func (s *Store) Create(in CreateCheckInput) (Check, error) {
	if err := validateCreateInput(in); err != nil {
		return Check{}, err
	}

	now := time.Now().UTC()
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}

	id, err := newID()
	if err != nil {
		return Check{}, err
	}

	check := Check{
		ID:               id,
		Name:             strings.TrimSpace(in.Name),
		URL:              strings.TrimSpace(in.URL),
		Method:           normalizeMethod(in.Method),
		IntervalSec:      defaultInt(in.IntervalSec, 60),
		TimeoutSec:       defaultInt(in.TimeoutSec, 5),
		ExpectedStatus:   defaultInt(in.ExpectedStatus, 200),
		NotifyOnRecovery: in.NotifyOnRecovery,
		Enabled:          enabled,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.mu.Lock()
	s.checks[check.ID] = check
	s.mu.Unlock()

	if err := s.save(); err != nil {
		return Check{}, err
	}

	return check, nil
}

func (s *Store) Update(id string, in UpdateCheckInput) (Check, error) {
	s.mu.Lock()
	check, ok := s.checks[id]
	if !ok {
		s.mu.Unlock()
		return Check{}, ErrNotFound
	}

	if in.Name != nil {
		v := strings.TrimSpace(*in.Name)
		if v == "" {
			s.mu.Unlock()
			return Check{}, errors.New("name cannot be empty")
		}
		check.Name = v
	}
	if in.URL != nil {
		v := strings.TrimSpace(*in.URL)
		if err := validateURL(v); err != nil {
			s.mu.Unlock()
			return Check{}, err
		}
		check.URL = v
	}
	if in.Method != nil {
		m := normalizeMethod(*in.Method)
		if m != MethodGet && m != MethodHead {
			s.mu.Unlock()
			return Check{}, errors.New("method must be GET or HEAD")
		}
		check.Method = m
	}
	if in.IntervalSec != nil {
		if *in.IntervalSec < 5 {
			s.mu.Unlock()
			return Check{}, errors.New("intervalSec must be >= 5")
		}
		check.IntervalSec = *in.IntervalSec
	}
	if in.TimeoutSec != nil {
		if *in.TimeoutSec < 1 {
			s.mu.Unlock()
			return Check{}, errors.New("timeoutSec must be >= 1")
		}
		check.TimeoutSec = *in.TimeoutSec
	}
	if in.ExpectedStatus != nil {
		if *in.ExpectedStatus < 100 || *in.ExpectedStatus > 599 {
			s.mu.Unlock()
			return Check{}, errors.New("expectedStatus must be 100..599")
		}
		check.ExpectedStatus = *in.ExpectedStatus
	}
	if in.NotifyOnRecovery != nil {
		check.NotifyOnRecovery = *in.NotifyOnRecovery
	}
	if in.Enabled != nil {
		check.Enabled = *in.Enabled
	}

	check.UpdatedAt = time.Now().UTC()
	s.checks[id] = check
	s.mu.Unlock()

	if err := s.save(); err != nil {
		return Check{}, err
	}

	return check, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	if _, ok := s.checks[id]; !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	delete(s.checks, id)
	s.mu.Unlock()

	return s.save()
}

func (s *Store) SetLastResult(id string, result CheckResult) error {
	s.mu.Lock()
	check, ok := s.checks[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	check.LastResult = &result
	check.UpdatedAt = time.Now().UTC()
	s.checks[id] = check
	s.mu.Unlock()

	return s.save()
}

func (s *Store) load() error {
	if s.dataFile == "" {
		return nil
	}

	raw, err := os.ReadFile(s.dataFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read data file: %w", err)
	}

	var items []Check
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("failed to parse data file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range items {
		s.checks[c.ID] = c
	}
	return nil
}

func (s *Store) save() error {
	if s.dataFile == "" {
		return nil
	}

	s.mu.RLock()
	items := make([]Check, 0, len(s.checks))
	for _, c := range s.checks {
		items = append(items, c)
	}
	s.mu.RUnlock()

	raw, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode data file: %w", err)
	}

	dir := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	if err := os.WriteFile(s.dataFile, raw, 0o600); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	return nil
}

func validateCreateInput(in CreateCheckInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	if err := validateURL(strings.TrimSpace(in.URL)); err != nil {
		return err
	}

	method := normalizeMethod(in.Method)
	if method != MethodGet && method != MethodHead {
		return errors.New("method must be GET or HEAD")
	}

	if in.IntervalSec != 0 && in.IntervalSec < 5 {
		return errors.New("intervalSec must be >= 5")
	}
	if in.TimeoutSec != 0 && in.TimeoutSec < 1 {
		return errors.New("timeoutSec must be >= 1")
	}
	expected := defaultInt(in.ExpectedStatus, 200)
	if expected < 100 || expected > 599 {
		return errors.New("expectedStatus must be 100..599")
	}

	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url must start with http:// or https://")
	}
	if u.Host == "" {
		return errors.New("url host is required")
	}
	return nil
}

func defaultInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func normalizeMethod(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return MethodGet
	}
	return v
}

func newID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
