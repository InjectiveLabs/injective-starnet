package pulumi

import (
	"errors"
	"fmt"

	"github.com/InjectiveLabs/injective-starnet/pkg/exec"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	INJECTIVED_HOME = "/home/injectived/artifacts"
)

func getStarnetKeyPath() (string, error) {
	return GetStarnetKey()
}

func SyncNodes(ctx *pulumi.Context, cfg Config, nodes Nodes, instances []*compute.Instance) error {
	// Sync validators
	for i, node := range nodes.Validators {
		sourcePath := fmt.Sprintf("%s/validators/%d/*", CHAIN_STRESSER_PATH, i)
		destPath := fmt.Sprintf("injectived@%s:%s", node.IP, INJECTIVED_HOME)

		// Get the starnet key path
		starnetKey, err := getStarnetKeyPath()
		if err != nil {
			return fmt.Errorf("failed to get starnet key: %w", err)
		}

		// Use local command to run scp
		_, err = local.NewCommand(ctx, fmt.Sprintf("copy-validator-%d", i), &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf("scp -r -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s -r %s %s", starnetKey, sourcePath, destPath)),
			Dir:    pulumi.String("."), // Run from current directory
		}, pulumi.DependsOn([]pulumi.Resource{instances[i]}))
		if err != nil {
			return fmt.Errorf("failed to create copy command for validator %d: %w", i, err)
		}
	}

	// Sync sentry nodes
	for i, node := range nodes.Sentries {
		sourcePath := fmt.Sprintf("%s/sentry-nodes/%d/*", CHAIN_STRESSER_PATH, i)
		destPath := fmt.Sprintf("injectived@%s:%s", node.IP, INJECTIVED_HOME)

		// Get the starnet key path
		starnetKey, err := getStarnetKeyPath()
		if err != nil {
			return fmt.Errorf("failed to get starnet key: %w", err)
		}

		// Use local command to run scp
		_, err = local.NewCommand(ctx, fmt.Sprintf("copy-sentry-%d", i), &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf("scp -r -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s -r %s %s", starnetKey, sourcePath, destPath)),
			Dir:    pulumi.String("."), // Run from current directory
		}, pulumi.DependsOn([]pulumi.Resource{instances[i]}))
		if err != nil {
			return fmt.Errorf("failed to create copy command for sentry %d: %w", i, err)
		}
	}

	// Install binaries on validators and sentries
	validatorIPs := make([]string, len(nodes.Validators))
	sentryIPs := make([]string, len(nodes.Sentries))
	for i, validator := range nodes.Validators {
		validatorIPs[i] = validator.IP
	}
	for i, sentry := range nodes.Sentries {
		sentryIPs[i] = sentry.IP
	}
	allIPs := append(validatorIPs, sentryIPs...)
	cmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -A -S none' git clone %s && cd injective-core && git checkout %s && make install", cfg.InjectiveConfig.Repository, cfg.InjectiveConfig.Branch)
	// Create an executor
	executor := &exec.Executor{
		Hosts:    allIPs,
		Cmd:      cmd,
		Username: "injectived",
	}

	// Execute the command on all hosts
	results := executor.Execute()

	// Process the results
	var errs []error
	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("Error on %s: %v\n", result.Host, result.Err)
			errs = append(errs, result.Err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
