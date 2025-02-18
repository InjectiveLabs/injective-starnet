package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CustomCommand represents a function as a Pulumi resource.
type CustomCommand struct {
	pulumi.ResourceState
}

// NewCustomCommand executes a function and registers it as a Pulumi resource.
func NewCustomCommand(ctx *pulumi.Context, name string, action func() error, dependsOn []pulumi.Resource, opts ...pulumi.ResourceOption) (*CustomCommand, error) {
	cmd := &CustomCommand{}

	// Add dependencies
	opts = append(opts, pulumi.DependsOn(dependsOn))

	err := ctx.RegisterComponentResource("starnet:chain:Command", name, cmd, opts...)
	if err != nil {
		return nil, err
	}

	// Execute the function
	if err := action(); err != nil {
		return nil, err
	}

	// Mark resource as successfully created
	ctx.RegisterResourceOutputs(cmd, pulumi.Map{})
	return cmd, nil
}
