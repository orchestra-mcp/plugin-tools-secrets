package store

import (
	"fmt"
	"math/rand"

	"github.com/orchestra-mcp/sdk-go/globaldb"
)

// Secret is an alias for globaldb.Secret.
type Secret = globaldb.Secret

// SecretStore provides CRUD operations on secrets via globaldb.
type SecretStore struct{}

// NewSecretStore creates a new SecretStore.
func NewSecretStore() (*SecretStore, error) {
	return &SecretStore{}, nil
}

// NewSecretID generates a secret ID in the format "SEC-XXXX".
func NewSecretID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 4)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "SEC-" + string(b)
}

// Create adds a new secret.
func (s *SecretStore) Create(sec *Secret) error {
	if _, err := globaldb.GetSecret(sec.ID); err == nil {
		return fmt.Errorf("secret %q already exists", sec.ID)
	}
	return globaldb.CreateSecret(sec)
}

// Get returns a single secret by ID with decrypted value.
func (s *SecretStore) Get(id string) (*Secret, error) {
	return globaldb.GetSecret(id)
}

// List returns all secrets (values masked).
func (s *SecretStore) List() ([]*Secret, error) {
	return globaldb.ListSecrets()
}

// ListByCategory returns secrets filtered by category (values masked).
func (s *SecretStore) ListByCategory(category string) ([]*Secret, error) {
	return globaldb.ListSecretsByCategory(category)
}

// Update modifies an existing secret via a mutation function.
func (s *SecretStore) Update(id string, fn func(sec *Secret)) error {
	return globaldb.UpdateSecret(id, fn)
}

// Delete removes a secret by ID.
func (s *SecretStore) Delete(id string) error {
	return globaldb.DeleteSecret(id)
}

// Search finds secrets by name, description, or tags.
func (s *SecretStore) Search(query string) ([]*Secret, error) {
	return globaldb.SearchSecrets(query)
}

// GetEnv returns secrets as key=value pairs for a given scope.
func (s *SecretStore) GetEnv(scope string) (map[string]string, error) {
	return globaldb.GetSecretEnv(scope)
}

// ImportEnv imports secrets from .env file content.
func (s *SecretStore) ImportEnv(data []byte, category, scope string) (int, error) {
	return globaldb.MigrateEnvFile(data, category, scope)
}

// MaskValue masks a secret value for display.
func MaskValue(v string) string {
	if len(v) <= 11 {
		return "****"
	}
	return v[:4] + "..." + v[len(v)-4:]
}
