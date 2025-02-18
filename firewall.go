package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createFirewall creates a firewall rule for all node types by checking the labels
func createFirewall(ctx *pulumi.Context, cfg Config, instances []*compute.Instance) error {
	// Collect all unique ports from both validators and sentries
	portsMap := make(map[string]struct{})

	// Add ports from validators
	for _, port := range cfg.Validators.NodePorts {
		portsMap[port] = struct{}{}
	}

	// Add ports from sentries
	for _, port := range cfg.Sentries.NodePorts {
		portsMap[port] = struct{}{}
	}

	// Convert to pulumi.StringArray
	var ports pulumi.StringArray
	for port := range portsMap {
		ports = append(ports, pulumi.String(port))
	}

	// Convert instances to []pulumi.Resource
	resources := make([]pulumi.Resource, len(instances))
	for i, instance := range instances {
		resources[i] = instance
	}

	_, err := compute.NewFirewall(ctx, "starnet-firewall", &compute.FirewallArgs{
		Network: pulumi.String("default"),
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("tcp"),
				Ports:    ports,
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"), // Adjust to restrict access if needed
		},
		TargetTags: pulumi.StringArray{pulumi.String("starnet-validator")},
	},
		pulumi.DependsOn(resources))
	if err != nil {
		return err
	}

	return nil
}
