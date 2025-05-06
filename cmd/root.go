package cmd

import (
	"github.com/InjectiveLabs/injective-starnet/cmd/pulumi"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "injective-starnet",
	Short: "Injective Starnet CLI",
	Long: `A CLI tool for managing the Injective Starnet network.
This includes deploying, managing, and monitoring the network infrastructure.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(pulumi.NewPulumiCmd())
}
