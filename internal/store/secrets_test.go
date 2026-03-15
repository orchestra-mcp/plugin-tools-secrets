package store

import (
	"testing"

	"github.com/orchestra-mcp/sdk-go/globaldb"
)

func TestNewSecretID(t *testing.T) {
	id := NewSecretID()
	if len(id) != 8 { // "SEC-" + 4 letters
		t.Fatalf("expected 8 chars, got %d: %s", len(id), id)
	}
	if id[:4] != "SEC-" {
		t.Fatalf("expected SEC- prefix, got %s", id)
	}
}

func TestSecretStoreCRUD(t *testing.T) {
	// Ensure globaldb is initialized.
	if _, err := globaldb.DB(); err != nil {
		t.Fatalf("init globaldb: %v", err)
	}
	defer globaldb.Close()

	store, err := NewSecretStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	// Create
	sec := &Secret{
		ID:          NewSecretID(),
		Name:        "TEST_API_KEY",
		Category:    "api_key",
		Value:       "sk-test-1234567890abcdef",
		Description: "Test API key",
		Tags:        []string{"test", "ci"},
		Scope:       "global",
	}
	if err := store.Create(sec); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get
	got, err := store.Get(sec.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "TEST_API_KEY" {
		t.Errorf("name: got %q, want TEST_API_KEY", got.Name)
	}
	if got.Value != "sk-test-1234567890abcdef" {
		t.Errorf("value not decrypted correctly: got %q", got.Value)
	}
	if got.Category != "api_key" {
		t.Errorf("category: got %q, want api_key", got.Category)
	}

	// List
	secrets, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, s := range secrets {
		if s.ID == sec.ID {
			found = true
			if s.Value != "****" {
				t.Errorf("list should mask values, got %q", s.Value)
			}
		}
	}
	if !found {
		t.Error("created secret not found in list")
	}

	// Update
	if err := store.Update(sec.ID, func(s *Secret) {
		s.Value = "sk-new-value-xyz"
		s.Description = "Updated description"
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	updated, _ := store.Get(sec.ID)
	if updated.Value != "sk-new-value-xyz" {
		t.Errorf("update value: got %q", updated.Value)
	}
	if updated.Description != "Updated description" {
		t.Errorf("update desc: got %q", updated.Description)
	}

	// Search
	results, err := store.Search("TEST_API")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("search found no results")
	}

	// GetEnv
	env, err := store.GetEnv("global")
	if err != nil {
		t.Fatalf("getenv: %v", err)
	}
	if env["TEST_API_KEY"] != "sk-new-value-xyz" {
		t.Errorf("getenv: got %q", env["TEST_API_KEY"])
	}

	// Delete
	if err := store.Delete(sec.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(sec.ID); err == nil {
		t.Error("expected error after delete")
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"short", "****"},
		{"12345678901", "****"},               // 11 chars
		{"sk-test-1234567890abcdef", "sk-t...cdef"}, // 24 chars
	}
	for _, tt := range tests {
		got := MaskValue(tt.input)
		if got != tt.want {
			t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestImportEnv(t *testing.T) {
	if _, err := globaldb.DB(); err != nil {
		t.Fatalf("init globaldb: %v", err)
	}
	defer globaldb.Close()

	store, _ := NewSecretStore()

	envContent := `# Database
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD=mysecret123

# Empty and comments
# SKIP_ME=value
`
	count, err := store.ImportEnv([]byte(envContent), "env", "staging")
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if count != 3 {
		t.Errorf("imported %d, want 3", count)
	}
}
