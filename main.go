package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

		// Create a single firewall for all nodes
		allInstances := []*compute.Instance{}

		// Spin up validators fleet
		validatorInstances, err := provisionNodes(ctx, cfg, VALIDATORS_TYPE)
		if err != nil {
			return err
		}
		allInstances = append(allInstances, validatorInstances...)

		// Spin up sentry nodes fleet
		sentryInstances, err := provisionNodes(ctx, cfg, SENTRIES_TYPE)
		if err != nil {
			return err
		}
		allInstances = append(allInstances, sentryInstances...)

		return nil
	})
}

func provisionNodes(ctx *pulumi.Context, cfg Config, nodeType string) ([]*compute.Instance, error) {
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
		return nil, fmt.Errorf("invalid node type: %s", nodeType)
	}

	if nodeType == VALIDATORS_TYPE {
		nodePoolSize = cfg.Validators.NodePoolSize
	} else if nodeType == SENTRIES_TYPE {
		nodePoolSize = cfg.Sentries.NodePoolSize
	} else {
		return nil, fmt.Errorf("invalid node type: %s", nodeType)
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
			return nil, err
		}

		// Extract IP for export
		ip := vm.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) *string {
			if len(nis) > 0 && len(nis[0].AccessConfigs) > 0 {
				return nis[0].AccessConfigs[0].NatIp
			}
			return nil
		}).(pulumi.StringPtrOutput)

		// Export hostname and IP
		ctx.Export(hostname, ip)

		// Add instance to the list
		instances = append(instances, vm)
	}

	// Create a firewall rule for all node types
	if err := CreateFirewall(ctx, cfg, instances, nodeType); err != nil {
		return nil, err
	}

	// Create a Nodes struct for configuration generation
	nodesStruct := Nodes{}
	if nodeType == VALIDATORS_TYPE {
		nodesStruct.Validators = make([]Node, nodePoolSize)
	} else if nodeType == SENTRIES_TYPE {
		nodesStruct.Sentries = make([]Node, nodePoolSize)
	}

	// Convert instances to interface slice for pulumi.All
	instanceInterfaces := make([]interface{}, len(instances))
	for i, instance := range instances {
		instanceInterfaces[i] = instance
	}

	// Wait for all instances to be created and collect their information
	pulumi.All(instanceInterfaces...).ApplyT(func(vms []interface{}) error {
		// Sort instances by index (which is already in order from creation)
		sort.Slice(vms, func(i, j int) bool {
			vm1 := vms[i].(*compute.Instance)
			vm2 := vms[j].(*compute.Instance)

			// Get indices from labels we set during creation
			var index1, index2 string

			vm1.Labels.ApplyT(func(labels map[string]string) error {
				index1 = labels["index"]
				return nil
			})

			vm2.Labels.ApplyT(func(labels map[string]string) error {
				index2 = labels["index"]
				return nil
			})

			return index1 < index2
		})

		// Create a slice to hold all the node information promises
		nodePromises := make([]pulumi.Output, len(vms))

		// Collect node information using values we already set
		for i, vmInterface := range vms {
			vm := vmInterface.(*compute.Instance)

			// Create a promise that will resolve to a Node
			nodePromise := pulumi.All(vm.Hostname, vm.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()).ApplyT(
				func(args []interface{}) (Node, error) {
					hostname := args[0].(*string)
					ip := args[1].(*string)

					if hostname == nil || ip == nil {
						return Node{}, fmt.Errorf("failed to get hostname or IP for node")
					}

					return Node{
						Host: *hostname,
						IP:   *ip,
					}, nil
				},
			)

			nodePromises[i] = nodePromise
		}

		// Convert nodePromises to interface slice for pulumi.All
		nodePromiseInterfaces := make([]interface{}, len(nodePromises))
		for i, promise := range nodePromises {
			nodePromiseInterfaces[i] = promise
		}

		// Wait for all node promises to resolve
		pulumi.All(nodePromiseInterfaces...).ApplyT(func(nodes []interface{}) error {
			for i, nodeInterface := range nodes {
				node := nodeInterface.(Node)

				// Add to appropriate slice
				if nodeType == VALIDATORS_TYPE {
					nodesStruct.Validators[i] = node
				} else if nodeType == SENTRIES_TYPE {
					nodesStruct.Sentries[i] = node
				}
			}

			// Generate configs
			generateCmd, err := NewCustomCommand(ctx, "generate-configs-"+nodeType, func() error {
				return GenerateNodesConfigs(cfg, nodesStruct, nodeType)
			}, []pulumi.Resource{})
			if err != nil {
				ctx.Log.Error(fmt.Sprintf("error generating configs %s: %v", nodeType, err), nil)
				return err
			}

			// Sync nodes
			_, err = NewCustomCommand(ctx, "copy-configs-"+nodeType, func() error {
				// Tmp Sleep to wait for SSH service to be fully available
				time.Sleep(10 * time.Second)
				return SyncNodes(ctx, cfg, nodesStruct, instances)
			}, []pulumi.Resource{generateCmd}) // Wait for generateCmd to finish before copying configs
			if err != nil {
				ctx.Log.Error(fmt.Sprintf("error syncing nodes %s: %v", nodeType, err), nil)
				return err
			}

			return nil
		})

		return nil
	})

	return instances, nil
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

func formatSSHKeys(keys SSHKeys) string {
	var formattedKeys string
	for _, k := range keys.Keys {
		formattedKeys += fmt.Sprintf("%s:%s %s\n", k.Username, k.Key, k.Username)
	}
	return formattedKeys
}
