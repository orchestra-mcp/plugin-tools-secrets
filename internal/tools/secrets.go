package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToolHandler is the standard tool handler function signature.
type ToolHandler = func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error)

// ---------- create_secret ----------

func CreateSecretSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Secret name / key (e.g. ANTHROPIC_API_KEY, DATABASE_URL)",
			},
			"value": map[string]any{
				"type":        "string",
				"description": "Secret value (will be encrypted at rest)",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "Category for grouping (default: general)",
				"enum":        []any{"general", "api_key", "token", "env", "database", "certificate", "password", "ssh", "webhook"},
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Human-readable description of what this secret is for",
			},
			"tags": map[string]any{
				"type":        "string",
				"description": "Comma-separated tags for filtering (e.g. production,stripe)",
			},
			"scope": map[string]any{
				"type":        "string",
				"description": "Scope: global (all workspaces) or a workspace name (default: global)",
			},
		},
		"required": []any{"name", "value"},
	})
	return s
}

func CreateSecret(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name", "value"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		value := helpers.GetString(req.Arguments, "value")
		category := helpers.GetString(req.Arguments, "category")
		desc := helpers.GetString(req.Arguments, "description")
		tagsStr := helpers.GetString(req.Arguments, "tags")
		scope := helpers.GetString(req.Arguments, "scope")

		if category == "" {
			category = "general"
		}
		if scope == "" {
			scope = "global"
		}

		var tags []string
		if tagsStr != "" {
			for _, t := range strings.Split(tagsStr, ",") {
				if trimmed := strings.TrimSpace(t); trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
		}

		sec := &store.Secret{
			ID:          store.NewSecretID(),
			Name:        name,
			Value:       value,
			Category:    category,
			Description: desc,
			Tags:        tags,
			Scope:       scope,
		}

		if err := s.Create(sec); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		md := formatSecretMD(sec, "Created secret", true)
		return helpers.TextResult(md), nil
	}
}

// ---------- list_secrets ----------

func ListSecretsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"category": map[string]any{
				"type":        "string",
				"description": "Filter by category (optional)",
			},
		},
	})
	return s
}

func ListSecrets(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		category := helpers.GetString(req.Arguments, "category")

		var secrets []*store.Secret
		var err error
		if category != "" {
			secrets, err = s.ListByCategory(category)
		} else {
			secrets, err = s.List()
		}
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		if len(secrets) == 0 {
			return helpers.TextResult("## Secrets\n\nNo secrets stored.\n"), nil
		}

		sort.Slice(secrets, func(i, j int) bool {
			return secrets[i].Name < secrets[j].Name
		})

		var b strings.Builder
		fmt.Fprintf(&b, "## Secrets (%d)\n\n", len(secrets))
		fmt.Fprintf(&b, "| ID | Name | Category | Scope | Tags | Description |\n")
		fmt.Fprintf(&b, "|----|------|----------|-------|------|-------------|\n")
		for _, sec := range secrets {
			tags := "-"
			if len(sec.Tags) > 0 {
				tags = strings.Join(sec.Tags, ", ")
			}
			desc := sec.Description
			if desc == "" {
				desc = "-"
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
				sec.ID, sec.Name, sec.Category, sec.Scope, tags, desc)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// ---------- get_secret ----------

func GetSecretSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"secret_id": map[string]any{
				"type":        "string",
				"description": "Secret ID (e.g. SEC-XXXX)",
			},
		},
		"required": []any{"secret_id"},
	})
	return s
}

func GetSecret(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "secret_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		id := helpers.GetString(req.Arguments, "secret_id")
		sec, err := s.Get(id)
		if err != nil {
			return helpers.ErrorResult("not_found", err.Error()), nil
		}

		md := formatSecretMD(sec, "Secret details", false)
		return helpers.TextResult(md), nil
	}
}

// ---------- update_secret ----------

func UpdateSecretSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"secret_id": map[string]any{
				"type":        "string",
				"description": "Secret ID to update",
			},
			"value": map[string]any{
				"type":        "string",
				"description": "New secret value (encrypted at rest)",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "New name/key",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "New category",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New description",
			},
			"tags": map[string]any{
				"type":        "string",
				"description": "New comma-separated tags (replaces existing)",
			},
			"scope": map[string]any{
				"type":        "string",
				"description": "New scope",
			},
		},
		"required": []any{"secret_id"},
	})
	return s
}

func UpdateSecret(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "secret_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		id := helpers.GetString(req.Arguments, "secret_id")
		newValue := helpers.GetString(req.Arguments, "value")
		newName := helpers.GetString(req.Arguments, "name")
		newCategory := helpers.GetString(req.Arguments, "category")
		newDesc := helpers.GetString(req.Arguments, "description")
		newTags := helpers.GetString(req.Arguments, "tags")
		newScope := helpers.GetString(req.Arguments, "scope")

		err := s.Update(id, func(sec *store.Secret) {
			if newValue != "" {
				sec.Value = newValue
			}
			if newName != "" {
				sec.Name = newName
			}
			if newCategory != "" {
				sec.Category = newCategory
			}
			if newDesc != "" {
				sec.Description = newDesc
			}
			if newTags != "" {
				var tags []string
				for _, t := range strings.Split(newTags, ",") {
					if trimmed := strings.TrimSpace(t); trimmed != "" {
						tags = append(tags, trimmed)
					}
				}
				sec.Tags = tags
			}
			if newScope != "" {
				sec.Scope = newScope
			}
		})
		if err != nil {
			return helpers.ErrorResult("not_found", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Updated secret **%s**", id)), nil
	}
}

// ---------- delete_secret ----------

func DeleteSecretSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"secret_id": map[string]any{
				"type":        "string",
				"description": "Secret ID to delete",
			},
		},
		"required": []any{"secret_id"},
	})
	return s
}

func DeleteSecret(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "secret_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		id := helpers.GetString(req.Arguments, "secret_id")
		if err := s.Delete(id); err != nil {
			return helpers.ErrorResult("not_found", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Deleted secret **%s**", id)), nil
	}
}

// ---------- search_secrets ----------

func SearchSecretsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query (matches name, description, tags)",
			},
		},
		"required": []any{"query"},
	})
	return s
}

func SearchSecrets(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "query"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		query := helpers.GetString(req.Arguments, "query")
		secrets, err := s.Search(query)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		if len(secrets) == 0 {
			return helpers.TextResult(fmt.Sprintf("No secrets matching %q", query)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Search Results (%d)\n\n", len(secrets))
		fmt.Fprintf(&b, "| ID | Name | Category | Scope | Description |\n")
		fmt.Fprintf(&b, "|----|------|----------|-------|-------------|\n")
		for _, sec := range secrets {
			desc := sec.Description
			if desc == "" {
				desc = "-"
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
				sec.ID, sec.Name, sec.Category, sec.Scope, desc)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// ---------- get_secret_env ----------

func GetSecretEnvSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"scope": map[string]any{
				"type":        "string",
				"description": "Filter by scope (default: all). Use 'global' for global secrets or a workspace name.",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "Output format: env (KEY=VALUE), json, or masked (default: env)",
				"enum":        []any{"env", "json", "masked"},
			},
		},
	})
	return s
}

func GetSecretEnv(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		scope := helpers.GetString(req.Arguments, "scope")
		format := helpers.GetString(req.Arguments, "format")
		if format == "" {
			format = "env"
		}

		env, err := s.GetEnv(scope)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		if len(env) == 0 {
			return helpers.TextResult("No secrets found for the specified scope."), nil
		}

		// Sort keys for consistent output.
		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var b strings.Builder
		switch format {
		case "json":
			data, _ := json.MarshalIndent(env, "", "  ")
			b.Write(data)
		case "masked":
			for _, k := range keys {
				fmt.Fprintf(&b, "%s=%s\n", k, store.MaskValue(env[k]))
			}
		default: // env
			for _, k := range keys {
				fmt.Fprintf(&b, "%s=%s\n", k, env[k])
			}
		}
		return helpers.TextResult(b.String()), nil
	}
}

// ---------- import_env ----------

func ImportEnvSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "Content of the .env file (KEY=VALUE lines, # comments ignored)",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "Category for all imported secrets (default: env)",
			},
			"scope": map[string]any{
				"type":        "string",
				"description": "Scope for all imported secrets (default: global)",
			},
		},
		"required": []any{"content"},
	})
	return s
}

func ImportEnv(s *store.SecretStore) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		content := helpers.GetString(req.Arguments, "content")
		category := helpers.GetString(req.Arguments, "category")
		scope := helpers.GetString(req.Arguments, "scope")

		count, err := s.ImportEnv([]byte(content), category, scope)
		if err != nil {
			return helpers.ErrorResult("import_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Imported **%d** secrets from .env content (category: %s, scope: %s)",
			count, orDefault(category, "env"), orDefault(scope, "global"))), nil
	}
}

// ---------- Helpers ----------

func formatSecretMD(sec *store.Secret, header string, maskValue bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "### %s: %s (%s)\n\n", header, sec.Name, sec.ID)
	fmt.Fprintf(&b, "- **Category:** %s\n", sec.Category)
	fmt.Fprintf(&b, "- **Scope:** %s\n", sec.Scope)
	if sec.Description != "" {
		fmt.Fprintf(&b, "- **Description:** %s\n", sec.Description)
	}
	if len(sec.Tags) > 0 {
		fmt.Fprintf(&b, "- **Tags:** %s\n", strings.Join(sec.Tags, ", "))
	}
	if maskValue {
		fmt.Fprintf(&b, "- **Value:** %s\n", store.MaskValue(sec.Value))
	} else {
		fmt.Fprintf(&b, "- **Value:** `%s`\n", sec.Value)
	}
	fmt.Fprintf(&b, "- **Created:** %s\n", sec.CreatedAt)
	fmt.Fprintf(&b, "- **Updated:** %s\n", sec.UpdatedAt)
	return b.String()
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
