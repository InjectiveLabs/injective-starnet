package pulumi

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var Main = func(ctx *pulumi.Context) error {

	cfg, err := LoadConfig(ctx)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Error loading configuration:%v", err.Error()), nil)
		return err
	}

	// Spin up validators fleet
	_, err = ProvisionNodes(ctx, cfg, VALIDATORS_TYPE)
	if err != nil {
		return err
	}
	// Spin up sentry nodes fleet
	_, err = ProvisionNodes(ctx, cfg, SENTRIES_TYPE)
	if err != nil {
		return err
	}

	return nil
}
