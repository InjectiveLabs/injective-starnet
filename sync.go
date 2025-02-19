package main

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	INJECTIVED_HOME = "/home/injectived/artifacts"
	STARNET_KEY     = "keys/starnet_key"
)

func syncNodes(ctx *pulumi.Context, nodes Nodes, instances []*compute.Instance) error {

	for i, validator := range nodes.Validators {
		sourcePath := fmt.Sprintf("%s/validators/%d/*", CHAIN_STRESSER_PATH, i)
		destPath := fmt.Sprintf("%s@%s:%s", "injectived", validator.IP, INJECTIVED_HOME)

		// Use local command to run scp
		_, err := local.NewCommand(ctx, fmt.Sprintf("copy-validator-%d", i), &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf("scp -r -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s -r %s %s", STARNET_KEY, sourcePath, destPath)),
			Dir:    pulumi.String("."), // Run from current directory
		}, pulumi.DependsOn([]pulumi.Resource{instances[i]}))

		if err != nil {
			ctx.Log.Error(fmt.Sprintf("failed to copy files to validator %d: %v", i, err), nil)
			return fmt.Errorf("failed to copy files to validator %d: %w", i, err)
		}
	}
	return nil
}
