package pulumi

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var Main = func(ctx *pulumi.Context) error {

	cfg, err := LoadConfig(ctx)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Error loading configuration:%v", err.Error()), nil)
		return err
	}
	// Create a single firewall for all nodes
	allInstances := []*compute.Instance{}

	// Spin up validators fleet
	validatorInstances, err := ProvisionNodes(ctx, cfg, VALIDATORS_TYPE)
	if err != nil {
		return err
	}
	allInstances = append(allInstances, validatorInstances...)

	// Spin up sentry nodes fleet
	sentryInstances, err := ProvisionNodes(ctx, cfg, SENTRIES_TYPE)
	if err != nil {
		return err
	}
	allInstances = append(allInstances, sentryInstances...)

	return nil
}
