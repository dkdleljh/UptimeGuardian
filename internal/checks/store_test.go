package checks

import "testing"

func TestStoreCreateAndGet(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	item, err := store.Create(CreateCheckInput{
		Name:           "Google",
		URL:            "https://www.google.com",
		Method:         "GET",
		IntervalSec:    10,
		TimeoutSec:     3,
		ExpectedStatus: 200,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.Get(item.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != "Google" {
		t.Fatalf("unexpected name: %s", got.Name)
	}
}

func TestStoreRejectsBadURL(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	_, err = store.Create(CreateCheckInput{
		Name: "Bad",
		URL:  "ftp://example.com",
	})
	if err == nil {
		t.Fatalf("expected error for invalid url")
	}
}
