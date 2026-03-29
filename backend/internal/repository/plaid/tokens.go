package plaid

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// tokenRecord holds a single Plaid access token and its metadata.
type tokenRecord struct {
	AccessToken     string    `json:"accessToken"`
	ItemID          string    `json:"itemId"`
	InstitutionName string    `json:"institutionName"`
	ConnectedAt     time.Time `json:"connectedAt"`
}

// tokenStore persists Plaid access tokens to a JSON file.
// All methods are safe for concurrent use.
type tokenStore struct {
	mu   sync.RWMutex
	path string
	data map[string]tokenRecord // slot → record
}

func newTokenStore(path string) (*tokenStore, error) {
	ts := &tokenStore{path: path, data: make(map[string]tokenRecord)}
	if err := ts.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ts, nil
}

func (ts *tokenStore) get(slot string) (tokenRecord, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	r, ok := ts.data[slot]
	return r, ok
}

func (ts *tokenStore) set(slot string, rec tokenRecord) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.data[slot] = rec
	return ts.save()
}

func (ts *tokenStore) slots() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	out := make([]string, 0, len(ts.data))
	for k := range ts.data {
		out = append(out, k)
	}
	return out
}

func (ts *tokenStore) load() error {
	b, err := os.ReadFile(ts.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &ts.data)
}

func (ts *tokenStore) save() error {
	b, err := json.MarshalIndent(ts.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ts.path, b, 0600)
}
