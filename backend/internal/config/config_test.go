package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APEX_ADDR", "")
	t.Setenv("PORT", "")
	t.Setenv("POLYGON_API_KEY", "")
	t.Setenv("COINGECKO_API_KEY", "")
	t.Setenv("PLAID_CLIENT_ID", "")
	t.Setenv("PLAID_SECRET", "")
	t.Setenv("PLAID_ENV", "")

	cfg := Load()

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.PlaidEnv != "sandbox" {
		t.Errorf("PlaidEnv = %q, want sandbox", cfg.PlaidEnv)
	}
	if cfg.PolygonKey != "" {
		t.Errorf("PolygonKey = %q, want empty", cfg.PolygonKey)
	}
}

func TestLoad_APEXAddr(t *testing.T) {
	t.Setenv("APEX_ADDR", ":9090")
	t.Setenv("PORT", "")
	t.Setenv("PLAID_ENV", "")

	cfg := Load()
	if cfg.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", cfg.Addr)
	}
}

func TestLoad_PORT_TakesPrecedence(t *testing.T) {
	t.Setenv("APEX_ADDR", "")
	t.Setenv("PORT", "3000")
	t.Setenv("PLAID_ENV", "")

	cfg := Load()
	if cfg.Addr != ":3000" {
		t.Errorf("Addr = %q, want :3000", cfg.Addr)
	}
}

func TestLoad_APIKeys(t *testing.T) {
	t.Setenv("POLYGON_API_KEY", "poly-key")
	t.Setenv("COINGECKO_API_KEY", "cg-key")
	t.Setenv("PLAID_CLIENT_ID", "plaid-id")
	t.Setenv("PLAID_SECRET", "plaid-secret")
	t.Setenv("PLAID_ENV", "production")
	t.Setenv("APEX_ADDR", "")
	t.Setenv("PORT", "")

	cfg := Load()

	if cfg.PolygonKey != "poly-key" {
		t.Errorf("PolygonKey = %q, want poly-key", cfg.PolygonKey)
	}
	if cfg.CoinGeckoKey != "cg-key" {
		t.Errorf("CoinGeckoKey = %q, want cg-key", cfg.CoinGeckoKey)
	}
	if cfg.PlaidClientID != "plaid-id" {
		t.Errorf("PlaidClientID = %q, want plaid-id", cfg.PlaidClientID)
	}
	if cfg.PlaidSecret != "plaid-secret" {
		t.Errorf("PlaidSecret = %q, want plaid-secret", cfg.PlaidSecret)
	}
	if cfg.PlaidEnv != "production" {
		t.Errorf("PlaidEnv = %q, want production", cfg.PlaidEnv)
	}
}

func TestLoadDotEnv_ParsesKeyValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "TEST_KEY_FOO=bar\nTEST_KEY_BAZ=qux\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("TEST_KEY_FOO")
	os.Unsetenv("TEST_KEY_BAZ")
	t.Cleanup(func() {
		os.Unsetenv("TEST_KEY_FOO")
		os.Unsetenv("TEST_KEY_BAZ")
	})

	loadDotEnv(path)

	if got := os.Getenv("TEST_KEY_FOO"); got != "bar" {
		t.Errorf("TEST_KEY_FOO = %q, want bar", got)
	}
	if got := os.Getenv("TEST_KEY_BAZ"); got != "qux" {
		t.Errorf("TEST_KEY_BAZ = %q, want qux", got)
	}
}

func TestLoadDotEnv_StripsQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte(`TEST_QUOTED="hello world"`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("TEST_QUOTED")
	t.Cleanup(func() { os.Unsetenv("TEST_QUOTED") })

	loadDotEnv(path)

	if got := os.Getenv("TEST_QUOTED"); got != "hello world" {
		t.Errorf("TEST_QUOTED = %q, want \"hello world\"", got)
	}
}

func TestLoadDotEnv_IgnoresComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "# this is a comment\nTEST_NOT_COMMENT=value\n# another comment\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("TEST_NOT_COMMENT")
	t.Cleanup(func() { os.Unsetenv("TEST_NOT_COMMENT") })

	loadDotEnv(path)

	if got := os.Getenv("TEST_NOT_COMMENT"); got != "value" {
		t.Errorf("TEST_NOT_COMMENT = %q, want value", got)
	}
}

func TestLoadDotEnv_DoesNotOverrideExistingEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte("TEST_OVERRIDE_VAR=from-file\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_OVERRIDE_VAR", "from-env")

	loadDotEnv(path)

	if got := os.Getenv("TEST_OVERRIDE_VAR"); got != "from-env" {
		t.Errorf("TEST_OVERRIDE_VAR = %q, want from-env (should not be overridden)", got)
	}
}

func TestLoadDotEnv_NonExistentFileIsOK(t *testing.T) {
	// Should not panic or return an error.
	loadDotEnv("/nonexistent/.env")
}

func TestLoadDotEnv_UsesFirstExistingFile(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.env")
	second := filepath.Join(dir, "second.env")

	if err := os.WriteFile(first, []byte("TEST_FIRST_FILE=yes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("TEST_FIRST_FILE=no\n"), 0600); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("TEST_FIRST_FILE")
	t.Cleanup(func() { os.Unsetenv("TEST_FIRST_FILE") })

	loadDotEnv(first, second)

	if got := os.Getenv("TEST_FIRST_FILE"); got != "yes" {
		t.Errorf("TEST_FIRST_FILE = %q, want yes", got)
	}
}
