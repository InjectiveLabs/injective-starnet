package pulumi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/InjectiveLabs/injective-starnet/pkg/pulumi"
	injectivepulumi "github.com/InjectiveLabs/injective-starnet/pkg/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func runWithSpinner(operation string, fn func() error) error {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	done := make(chan error, 1)

	// Start the operation in a goroutine
	go func() {
		done <- fn()
	}()

	// Show spinner while waiting
	for {
		select {
		case err := <-done:
			fmt.Printf("\r%s %s\n", spinner[0], operation)
			return err
		default:
			fmt.Printf("\r%s %s", spinner[i], operation)
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// setupPulumiStack creates a temporary workspace and stack for Pulumi operations
func setupPulumiStack(ctx context.Context, validatorSize, sentrySize int, buildBranch, artifactsPath string) (auto.Stack, error) {
	// Set artifacts path if provided
	if artifactsPath != "" {
		injectivepulumi.SetArtifactsPath(artifactsPath)
		fmt.Printf("Setting artifacts path to: %s\n", artifactsPath)
	}

	// Create a temporary directory for Pulumi.yaml
	tempDir, err := os.MkdirTemp("", "pulumi-config-*")
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Note: We don't delete the temp directory here as Pulumi needs it

	// Get embedded Pulumi.yaml content
	yamlContent, err := injectivepulumi.GetPulumiYAML()
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to get embedded Pulumi.yaml: %w", err)
	}

	// Write Pulumi.yaml to temp directory
	if err := os.WriteFile(filepath.Join(tempDir, "Pulumi.yaml"), yamlContent, 0644); err != nil {
		return auto.Stack{}, fmt.Errorf("failed to write Pulumi.yaml: %w", err)
	}

	// Load embedded config
	configMap, err := injectivepulumi.LoadEmbeddedConfig()
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to load embedded config: %w", err)
	}

	// Update branch if specified
	if buildBranch != "" {
		if injectiveConfig, ok := configMap["injective"].(map[string]interface{}); ok {
			injectiveConfig["branch"] = buildBranch
			configMap["injective"] = injectiveConfig
		}
	}

	// Update node pool sizes if specified
	if validatorSize > 0 {
		// Find and update the validator node pool size
		if nodePools, ok := configMap["nodePools"].([]interface{}); ok && len(nodePools) > 0 {
			if validatorPool, ok := nodePools[0].(map[string]interface{}); ok {
				validatorPool["nodePoolSize"] = validatorSize
			}
		}
	}

	if sentrySize > 0 {
		// Find and update the sentry node pool size
		if nodePools, ok := configMap["nodePools"].([]interface{}); ok && len(nodePools) > 1 {
			if sentryPool, ok := nodePools[1].(map[string]interface{}); ok {
				sentryPool["nodePoolSize"] = sentrySize
			}
		}
	}

	// Initialize project settings
	desc := "Injective Starnet Infrastructure"
	projectSettings := workspace.Project{
		Name:        "injective-starnet",
		Runtime:     workspace.NewProjectRuntimeInfo("go", nil),
		Description: &desc,
		Main:        ".",
	}

	// Create workspace with project settings and program
	workspace, err := auto.NewLocalWorkspace(ctx,
		auto.Project(projectSettings),
		auto.Program(pulumi.Main),
		auto.WorkDir(tempDir),
	)
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create or select the stack
	stack, err := auto.UpsertStack(ctx, "starnet", workspace)
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to create/select stack: %w", err)
	}

	// Set all config values from the updated config
	for key, value := range configMap {
		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%v", v)
		default:
			// For complex types like arrays and maps, convert to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return auto.Stack{}, fmt.Errorf("failed to marshal config value for %s: %w", key, err)
			}
			strValue = string(jsonBytes)
		}

		err = stack.SetConfig(ctx, key, auto.ConfigValue{
			Value: strValue,
		})
		if err != nil {
			return auto.Stack{}, fmt.Errorf("failed to set config %s: %w", key, err)
		}
	}

	return stack, nil
}
