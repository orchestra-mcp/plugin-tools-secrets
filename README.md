# Orchestra Plugin: tools-secrets

Encrypted secrets management for [Orchestra MCP](https://github.com/orchestra-mcp/framework). Store API keys, tokens, passwords, and .env variables with AES-256-GCM encryption at rest. Secrets are shared across all workspaces via `~/.orchestra/db/global.db`.

## Install

```bash
go install github.com/orchestra-mcp/plugin-tools-secrets/cmd@latest
```

Or build from source:

```bash
make build-tools-secrets
```

## Usage

The plugin is bundled in-process with the Orchestra CLI. No extra configuration needed — secrets tools are available immediately after `orchestra serve`.

For standalone use, add to `plugins.yaml`:

```yaml
- id: tools.secrets
  binary: ./bin/tools-secrets
  enabled: true
```

## Tools (8)

| Tool | Description |
|------|-------------|
| `create_secret` | Create a new encrypted secret (API key, token, password, .env variable) |
| `list_secrets` | List all stored secrets with masked values, optionally filtered by category |
| `get_secret` | Get a secret's full details including its decrypted value |
| `update_secret` | Update a secret's value, name, category, description, tags, or scope |
| `delete_secret` | Delete a secret permanently |
| `search_secrets` | Search secrets by name, description, or tags |
| `get_secret_env` | Export secrets as KEY=VALUE (.env), JSON, or masked output |
| `import_env` | Import secrets from .env file content |

## Categories

Secrets can be organized into categories: `general`, `api_key`, `token`, `env`, `database`, `certificate`, `password`, `ssh`, `webhook`.

## Scopes

Each secret has a scope (default: `global`). Use scopes to organize secrets by environment (production, staging) or workspace name.

## Security

- **Encryption**: AES-256-GCM encryption at rest using a machine-local key at `~/.orchestra/db/encryption.key`
- **Storage**: SQLite in `~/.orchestra/db/global.db` — never synced via git
- **Masking**: List operations always return masked values; decryption only via `get_secret` or `get_secret_env`
- **Permissions**: Key file is 0600 (owner-only read/write)

## Related Packages

- [sdk-go](https://github.com/orchestra-mcp/sdk-go) — Plugin SDK
- [gen-go](https://github.com/orchestra-mcp/gen-go) — Generated Protobuf types
