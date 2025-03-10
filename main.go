package main

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func formatSSHKeys(keys SSHKeys) string {
	var formattedKeys string
	for _, k := range keys.Keys {
		formattedKeys += fmt.Sprintf("%s:%s %s\n", k.Username, k.Key, k.Username)
	}
	return formattedKeys
}

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		cfg, err := LoadConfig(ctx)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error loading configuration:%v", err.Error()), nil)
			return err
		}
		// Validate all artifacts are provided, fail early if not
		if err := checkBuildArtifacts(cfg); err != nil {
			ctx.Log.Error(fmt.Sprintf("Error checking build artifacts: %v", err.Error()), nil)
			return err
		}

		// Spin up validators fleet
		if err := provisionNodes(ctx, cfg, VALIDATORS_TYPE); err != nil {
			return err
		}

		// Spin up sentry nodes fleet
		if err := provisionNodes(ctx, cfg, SENTRIES_TYPE); err != nil {
			return err
		}

		return nil
	})
}

func provisionNodes(ctx *pulumi.Context, cfg Config, nodeType string) error {
	var nodes Nodes
	instances := []*compute.Instance{}
	totalRegions := 0
	zonesPerRegion := 0
	regions := []string{}
	nodePoolSize := 0
	machineType := ""
	image := ""
	diskSizeGB := 0
	diskType := ""
	nodesAutoDelete := false

	if nodeType == VALIDATORS_TYPE {
		regions = cfg.Validators.NodeRegions
		totalRegions = len(regions)
		zonesPerRegion = cfg.Validators.NodeZonesPerRegion
		machineType = cfg.Validators.NodeMachineType
		image = cfg.Validators.NodeImage
		diskSizeGB = cfg.Validators.NodeDiskSizeGB
		diskType = cfg.Validators.NodeDiskType
		nodesAutoDelete = cfg.Validators.NodesAutoDelete
	} else if nodeType == SENTRIES_TYPE {
		regions = cfg.Sentries.NodeRegions
		totalRegions = len(regions)
		zonesPerRegion = cfg.Sentries.NodeZonesPerRegion
		machineType = cfg.Sentries.NodeMachineType
		image = cfg.Sentries.NodeImage
		diskSizeGB = cfg.Sentries.NodeDiskSizeGB
		diskType = cfg.Sentries.NodeDiskType
		nodesAutoDelete = cfg.Sentries.NodesAutoDelete
	} else {
		return fmt.Errorf("invalid node type: %s", nodeType)
	}

	if nodeType == VALIDATORS_TYPE {
		nodePoolSize = cfg.Validators.NodePoolSize
	} else if nodeType == SENTRIES_TYPE {
		nodePoolSize = cfg.Sentries.NodePoolSize
	} else {
		return fmt.Errorf("invalid node type: %s", nodeType)
	}

	// Create only nodePoolSize number of instances
	for instanceNum := range nodePoolSize {
		// Distribute across regions round-robin style
		regionIndex := instanceNum % totalRegions
		// Distribute across zones within the selected region
		zoneIndex := (instanceNum / totalRegions) % zonesPerRegion

		zone := fmt.Sprintf("%s-%c", regions[regionIndex], 'b'+rune(zoneIndex))
		name := fmt.Sprintf("starnet-%s-%d", nodeType, instanceNum)
		hostname := fmt.Sprintf("%s%s", name, ".injective.network")

		vm, err := compute.NewInstance(ctx, hostname, &compute.InstanceArgs{
			MachineType: pulumi.String(machineType),
			AdvancedMachineFeatures: &compute.InstanceAdvancedMachineFeaturesArgs{
				ThreadsPerCore: pulumi.Int(1),
			},

			Zone: pulumi.String(zone),
			BootDisk: &compute.InstanceBootDiskArgs{
				InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
					Image: pulumi.String(image),
					Size:  pulumi.Int(diskSizeGB),
					Type:  pulumi.String(diskType),
				},
				AutoDelete: pulumi.Bool(true), // Delete the disk when the VM is deleted, otherwise it will be retained and charged
			},
			// Configure 4 local NVMe SSDs for RAID0
			// Local disk are ephemeral by nature and cannot persist beyond the instance's lifetime.
			ScratchDisks: compute.InstanceScratchDiskArray{
				&compute.InstanceScratchDiskArgs{
					Interface: pulumi.String("NVME"),
				},
				&compute.InstanceScratchDiskArgs{
					Interface: pulumi.String("NVME"),
				},
				&compute.InstanceScratchDiskArgs{
					Interface: pulumi.String("NVME"),
				},
				&compute.InstanceScratchDiskArgs{
					Interface: pulumi.String("NVME"),
				},
			},
			NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
				&compute.InstanceNetworkInterfaceArgs{
					Network: pulumi.String("default"),
					AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
						&compute.InstanceNetworkInterfaceAccessConfigArgs{},
					},
				},
			},
			// Add machine startup script
			Metadata: pulumi.StringMap{
				"ssh-keys": pulumi.String(formatSSHKeys(cfg.SSHKeys)),
			},
			// Label the node, so we have unique identifiers for each node
			Labels: pulumi.StringMap{
				"type":  pulumi.String(nodeType),
				"index": pulumi.String(fmt.Sprintf("%d", instanceNum)),
			},
			Name:     pulumi.String(name),
			Hostname: pulumi.String(hostname),
			Tags:     pulumi.StringArray{pulumi.String(fmt.Sprintf("starnet-%s", nodeType))}, // Apply firewall rules
			// Set scheduling to preemptible if nodesAutoDelete is true (e.g auto delete after 24h)
			Scheduling: &compute.InstanceSchedulingArgs{
				Preemptible:      pulumi.Bool(nodesAutoDelete),
				AutomaticRestart: pulumi.Bool(false),
			},
		})
		if err != nil {
			return err
		}

		ip := vm.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) string {
			if len(nis) > 0 && len(nis[0].AccessConfigs) > 0 {
				return *nis[0].AccessConfigs[0].NatIp
			}
			return ""
		}).(pulumi.StringOutput)

		// Use ApplyT, its like defer to extract the actual IP and store it as a Go string
		// Since Pulumi outputs (pulumi.StringOutput) cannot be directly assigned to a Go string because Pulumi outputs are asynchronous.
		ip.ApplyT(func(ipStr string) string {
			if nodeType == VALIDATORS_TYPE {
				nodes.Validators = append(nodes.Validators, Node{
					Host: hostname,
					IP:   ipStr,
				})
			} else if nodeType == SENTRIES_TYPE {
				nodes.Sentries = append(nodes.Sentries, Node{
					Host: hostname,
					IP:   ipStr,
				})
			}
			return ipStr
		})

		instances = append(instances, vm)
		ctx.Export(hostname, ip)
		// Concatenate all IPs to a single string so we can easily pass it to Prometheus
		ips := ""
		if nodeType == VALIDATORS_TYPE {
			for _, ip := range nodes.Validators {
				ips += ip.IP + ","
			}
		} else if nodeType == SENTRIES_TYPE {
			for _, ip := range nodes.Sentries {
				ips += ip.IP + ","
			}
		}
		// Convert the concatenated string to a pulumi.String
		ctx.Export("ips", pulumi.String(ips))
	}

	// Create a firewall rule for all node types
	if err := CreateFirewall(ctx, cfg, instances, nodeType); err != nil {
		return err
	}

	// Convert instances to []pulumi.Resource and collect all IPs
	resources := make([]pulumi.Resource, len(instances))
	allIPs := make([]pulumi.Output, len(instances))
	for i, instance := range instances {
		resources[i] = instance
		allIPs[i] = instance.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) string {
			return *nis[0].AccessConfigs[0].NatIp
		}).(pulumi.StringOutput)
	}

	// Convert allIPs to []interface{}
	interfaceIPs := make([]any, len(allIPs))
	for i, ip := range allIPs {
		interfaceIPs[i] = ip
	}

	// Wait for all instance/IPs to be available before generating configs
	pulumi.All(interfaceIPs...).ApplyT(func(ips []any) error {
		generateCmd, err := NewCustomCommand(ctx, "generate-configs-"+nodeType, func() error {
			return GenerateNodesConfigs(cfg, nodes, nodeType)
		}, resources)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("error generating configs %s: %v", nodeType, err), nil)
			return err
		}

		// Sync nodes
		_, err = NewCustomCommand(ctx, "copy-configs-"+nodeType, func() error {
			// Tmp Sleep to wait for SSH service to be fully available
			time.Sleep(10 * time.Second)
			return SyncNodes(ctx, nodes, instances)
		}, []pulumi.Resource{generateCmd}) // Wait for generateCmd to finish before copying configs
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("error syncing nodes %s: %v", nodeType, err), nil)
			return err
		}
		return nil
	})

	return nil

}

func checkBuildArtifacts(cfg Config) error {
	// Fail early if artifacts are not provided, at this point we don't have any instances yet.
	nodesMock := Nodes{
		Validators: make([]Node, cfg.Validators.NodePoolSize),
		Sentries:   make([]Node, cfg.Sentries.NodePoolSize),
	}
	for i := range nodesMock.Validators {
		nodesMock.Validators[i] = Node{
			Host: fmt.Sprintf("starnet-validator-%d.injective.network", i),
			IP:   fmt.Sprintf("10.0.0.%d", i),
		}
	}
	for i := range nodesMock.Sentries {
		nodesMock.Sentries[i] = Node{
			Host: fmt.Sprintf("starnet-sentry-%d.injective.network", i),
			IP:   fmt.Sprintf("10.0.0.%d", i),
		}
	}
	if err := CheckArtifacts(CHAIN_STRESSER_PATH, nodesMock); err != nil {
		return fmt.Errorf("error checking artifacts: %v", err)
	}
	return nil
}
