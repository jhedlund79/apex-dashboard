package plaid

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewTokenStore_NonExistentFileIsOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent-tokens.json")

	ts, err := newTokenStore(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts == nil {
		t.Fatal("expected non-nil token store")
	}
}

func TestNewTokenStore_CorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	if err := os.WriteFile(path, []byte("not valid json {{{"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := newTokenStore(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestTokenStore_GetEmptyReturnsNotFound(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newTokenStore(filepath.Join(dir, "tokens.json"))

	_, ok := ts.get("principal")
	if ok {
		t.Error("expected ok=false for missing slot")
	}
}

func TestTokenStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newTokenStore(filepath.Join(dir, "tokens.json"))

	rec := tokenRecord{
		AccessToken:     "access-sandbox-abc",
		ItemID:          "item-123",
		InstitutionName: "Principal Financial",
		ConnectedAt:     time.Now().Truncate(time.Second),
	}

	if err := ts.set("principal", rec); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, ok := ts.get("principal")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got.AccessToken != rec.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, rec.AccessToken)
	}
	if got.ItemID != rec.ItemID {
		t.Errorf("ItemID = %q, want %q", got.ItemID, rec.ItemID)
	}
	if got.InstitutionName != rec.InstitutionName {
		t.Errorf("InstitutionName = %q, want %q", got.InstitutionName, rec.InstitutionName)
	}
}

func TestTokenStore_SetPersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	ts1, _ := newTokenStore(path)
	rec := tokenRecord{AccessToken: "access-token-xyz", ItemID: "item-456", InstitutionName: "Morgan Stanley"}
	if err := ts1.set("morganstanley", rec); err != nil {
		t.Fatalf("set: %v", err)
	}

	// Create a second store from the same file — it should read the persisted token.
	ts2, err := newTokenStore(path)
	if err != nil {
		t.Fatalf("newTokenStore (reload): %v", err)
	}

	got, ok := ts2.get("morganstanley")
	if !ok {
		t.Fatal("expected token after reload")
	}
	if got.AccessToken != "access-token-xyz" {
		t.Errorf("AccessToken = %q, want access-token-xyz", got.AccessToken)
	}
}

func TestTokenStore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	ts, _ := newTokenStore(path)
	if err := ts.set("principal", tokenRecord{AccessToken: "tok"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %04o, want 0600", perm)
	}
}

func TestTokenStore_Slots(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newTokenStore(filepath.Join(dir, "tokens.json"))

	if slots := ts.slots(); len(slots) != 0 {
		t.Errorf("slots on empty store = %v, want empty", slots)
	}

	ts.set("principal", tokenRecord{AccessToken: "tok1"})     //nolint
	ts.set("morganstanley", tokenRecord{AccessToken: "tok2"}) //nolint

	slots := ts.slots()
	if len(slots) != 2 {
		t.Errorf("len(slots) = %d, want 2", len(slots))
	}

	slotSet := map[string]bool{}
	for _, s := range slots {
		slotSet[s] = true
	}
	if !slotSet["principal"] || !slotSet["morganstanley"] {
		t.Errorf("slots = %v, want [principal morganstanley]", slots)
	}
}

func TestTokenStore_OverwriteSlot(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newTokenStore(filepath.Join(dir, "tokens.json"))

	ts.set("principal", tokenRecord{AccessToken: "old-token"})  //nolint
	ts.set("principal", tokenRecord{AccessToken: "new-token"})  //nolint

	got, _ := ts.get("principal")
	if got.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want new-token", got.AccessToken)
	}
	// Should still be one slot, not two.
	if len(ts.slots()) != 1 {
		t.Errorf("len(slots) = %d, want 1", len(ts.slots()))
	}
}

func TestTokenStore_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newTokenStore(filepath.Join(dir, "tokens.json"))

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ts.set("principal", tokenRecord{AccessToken: "tok"}) //nolint
		}()
		go func() {
			defer wg.Done()
			ts.get("principal")
		}()
	}
	wg.Wait()
	// No assertion needed — validates no data races when run with -race.
}
