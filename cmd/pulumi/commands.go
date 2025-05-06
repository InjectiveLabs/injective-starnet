package pulumi

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/spf13/cobra"
)

func NewPulumiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage the Injective Starnet network using Pulumi",
	}

	cmd.AddCommand(newUpCmd())
	cmd.AddCommand(newPreviewCmd())
	cmd.AddCommand(newDestroyCmd())

	return cmd
}

func newUpCmd() *cobra.Command {
	var (
		validatorSize int
		sentrySize    int
	)

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Deploy the Injective Starnet network",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Setup Pulumi stack with configuration
			stack, err := setupPulumiStack(ctx, validatorSize, sentrySize)
			if err != nil {
				return err
			}

			return runWithSpinner("Deploying resources...", func() error {
				res, err := stack.Up(ctx, optup.ProgressStreams(os.Stdout))
				if err != nil {
					return fmt.Errorf("pulumi up failed: %w", err)
				}

				// Print summary
				if res.Summary.ResourceChanges != nil {
					for opType, count := range *res.Summary.ResourceChanges {
						fmt.Printf("    %s: %d resources\n", opType, count)
					}
				}

				// Print outputs
				if len(res.Outputs) > 0 {
					for k, v := range res.Outputs {
						fmt.Printf("    %s: %v\n", k, v.Value)
					}
				}

				return nil
			})
		},
	}

	cmd.Flags().IntVar(&validatorSize, "validators", 0, "Override the number of validator nodes")
	cmd.Flags().IntVar(&sentrySize, "sentries", 0, "Override the number of sentry nodes")

	return cmd
}

func newPreviewCmd() *cobra.Command {
	var (
		validatorSize int
		sentrySize    int
	)

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview the Injective Starnet network deployment/changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Setup Pulumi stack with configuration
			stack, err := setupPulumiStack(ctx, validatorSize, sentrySize)
			if err != nil {
				return err
			}

			return runWithSpinner("Generating preview...", func() error {
				preview, err := stack.Preview(ctx, optpreview.ProgressStreams(os.Stdout))
				if err != nil {
					return fmt.Errorf("pulumi preview failed: %w", err)
				}

				for opType, count := range preview.ChangeSummary {
					fmt.Printf("    %s: %d resources\n", opType, count)
				}

				return nil
			})
		},
	}

	cmd.Flags().IntVar(&validatorSize, "validators", 0, "Override the number of validator nodes")
	cmd.Flags().IntVar(&sentrySize, "sentries", 0, "Override the number of sentry nodes")

	return cmd
}

func newDestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the Injective Starnet network",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Setup Pulumi stack with configuration
			stack, err := setupPulumiStack(ctx, 0, 0)
			if err != nil {
				return err
			}

			return runWithSpinner("Destroying resources...", func() error {
				res, err := stack.Destroy(ctx, optdestroy.ProgressStreams(os.Stdout))
				if err != nil {
					return fmt.Errorf("pulumi destroy failed: %w", err)
				}

				// Print summary
				if res.Summary.ResourceChanges != nil {
					for opType, count := range *res.Summary.ResourceChanges {
						fmt.Printf("    %s: %d resources\n", opType, count)
					}
				}

				return nil
			})
		},
	}

	return cmd
}
