# Tools Reference

## create_secret

Create a new encrypted secret.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Secret name/key (e.g. ANTHROPIC_API_KEY) |
| `value` | string | Yes | Secret value (encrypted at rest) |
| `category` | string | No | Category: general, api_key, token, env, database, certificate, password, ssh, webhook |
| `description` | string | No | What this secret is for |
| `tags` | string | No | Comma-separated tags |
| `scope` | string | No | Scope: global or workspace name (default: global) |

## list_secrets

List all secrets with masked values.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string | No | Filter by category |

## get_secret

Get secret details including decrypted value.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secret_id` | string | Yes | Secret ID (e.g. SEC-XXXX) |

## update_secret

Update a secret's fields.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secret_id` | string | Yes | Secret ID |
| `value` | string | No | New value |
| `name` | string | No | New name |
| `category` | string | No | New category |
| `description` | string | No | New description |
| `tags` | string | No | New tags (replaces existing) |
| `scope` | string | No | New scope |

## delete_secret

Delete a secret permanently.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secret_id` | string | Yes | Secret ID |

## search_secrets

Search secrets by name, description, or tags.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query |

## get_secret_env

Export secrets as environment variables.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scope` | string | No | Filter by scope (default: all) |
| `format` | string | No | Output: env, json, or masked (default: env) |

## import_env

Import secrets from .env content.

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `content` | string | Yes | .env file content (KEY=VALUE lines) |
| `category` | string | No | Category for imported secrets (default: env) |
| `scope` | string | No | Scope for imported secrets (default: global) |
