/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/user"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "REST API server",
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(
			fx.Provide(user.Module),
		)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
