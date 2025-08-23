package cmd

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity"
	shared "github.com/cristiano-pacheco/pingo/internal/shared/modules"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "REST API server",
	Run: func(_ *cobra.Command, _ []string) {
		app := fx.New(
			shared.Module,
			identity.Module,
		)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
