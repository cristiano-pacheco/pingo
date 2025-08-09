package main

import (
	"context"
	"log"

	"github.com/cristiano-pacheco/pingo/cmd"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
)

func main() {
	config.Init()
	cfg := config.GetConfig()
	otel.Init(cfg)

	defer func() {
		if err := otel.Trace().Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	cmd.Execute()
}
