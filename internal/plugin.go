package internal

import (
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal/store"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal/tools"
)

// SecretsPlugin holds the shared dependencies for all tool handlers.
type SecretsPlugin struct {
	Store *store.SecretStore
}

// RegisterTools registers all 8 secrets tools on the given plugin builder.
func (p *SecretsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	s := p.Store

	builder.RegisterTool("create_secret",
		"Create a new encrypted secret (API key, token, password, .env variable). Stored locally with AES-256-GCM encryption, shared across workspaces.",
		tools.CreateSecretSchema(), tools.CreateSecret(s))

	builder.RegisterTool("list_secrets",
		"List all stored secrets with masked values. Optionally filter by category.",
		tools.ListSecretsSchema(), tools.ListSecrets(s))

	builder.RegisterTool("get_secret",
		"Get a secret's full details including its decrypted value.",
		tools.GetSecretSchema(), tools.GetSecret(s))

	builder.RegisterTool("update_secret",
		"Update a secret's value, name, category, description, tags, or scope.",
		tools.UpdateSecretSchema(), tools.UpdateSecret(s))

	builder.RegisterTool("delete_secret",
		"Delete a secret permanently.",
		tools.DeleteSecretSchema(), tools.DeleteSecret(s))

	builder.RegisterTool("search_secrets",
		"Search secrets by name, description, or tags.",
		tools.SearchSecretsSchema(), tools.SearchSecrets(s))

	builder.RegisterTool("get_secret_env",
		"Export secrets as KEY=VALUE pairs (.env format), JSON, or masked output. Useful for injecting secrets into processes.",
		tools.GetSecretEnvSchema(), tools.GetSecretEnv(s))

	builder.RegisterTool("import_env",
		"Import secrets from .env file content. Each KEY=VALUE line becomes an encrypted secret.",
		tools.ImportEnvSchema(), tools.ImportEnv(s))
}
