package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal"
	"github.com/orchestra-mcp/plugin-tools-secrets/internal/store"
)

func main() {
	builder := plugin.New("tools.secrets").
		Version("0.1.0").
		Description("Encrypted secrets management — API keys, tokens, passwords, .env variables. AES-256-GCM at rest, shared across workspaces.").
		Author("Orchestra").
		Binary("tools-secrets")

	secretStore, err := store.NewSecretStore()
	if err != nil {
		log.Fatalf("tools.secrets: init store: %v", err)
	}

	sp := &internal.SecretsPlugin{Store: secretStore}
	sp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("tools.secrets: %v", err)
	}
}
