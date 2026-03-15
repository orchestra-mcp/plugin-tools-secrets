package toolssecrets

import (
	"github.com/orchestra-mcp/plugin-tools-secrets/internal"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal/store"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all 8 secrets tools to the builder.
func Register(builder *plugin.PluginBuilder) error {
	secretStore, err := store.NewSecretStore()
	if err != nil {
		return err
	}
	sp := &internal.SecretsPlugin{
		Store: secretStore,
	}
	sp.RegisterTools(builder)
	return nil
}
