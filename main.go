package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Node struct {
	Host          string `json:"host"`
	IP            string `json:"ip"`
	NetworkNodeID string `json:"NetworkNodeID"`
}

type Nodes struct {
	Validators []Node
	Sentries   []Node
}

func formatSSHKeys(keys SSHKeys) string {
	var formattedKeys string
	for _, k := range keys.Keys {
		formattedKeys += fmt.Sprintf("%s:%s %s\n", k.Username, k.Key, k.Username)
	}
	return formattedKeys
}

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {
		var nodes Nodes

		cfg, err := loadConfig(ctx)
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("Error loading configuration:%v", err.Error()), nil)
			return err
		}

		var instances []*compute.Instance

		// Spin up validators fleet
		regions := cfg.Validators.NodeRegions
		totalRegions := len(regions)

		// Create only nodePoolSize number of instances
		for instanceNum := 0; instanceNum < cfg.Validators.NodePoolSize; instanceNum++ {
			// Distribute across regions round-robin style
			regionIndex := instanceNum % totalRegions
			// Distribute across zones within the selected region
			zoneIndex := (instanceNum / totalRegions) % cfg.Validators.NodeZonesPerRegion

			zone := fmt.Sprintf("%s-%c", regions[regionIndex], 'b'+rune(zoneIndex))
			name := fmt.Sprintf("starnet-validator-%d", instanceNum)
			hostname := fmt.Sprintf("%s%s", name, ".injective.network")

			vm, err := compute.NewInstance(ctx, hostname, &compute.InstanceArgs{
				MachineType: pulumi.String(cfg.Validators.NodeMachineType),
				Zone:        pulumi.String(zone),
				BootDisk: &compute.InstanceBootDiskArgs{
					InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
						Image: pulumi.String(cfg.Validators.NodeImage),
						Size:  pulumi.Int(cfg.Validators.NodeDiskSizeGB),
						Type:  pulumi.String(cfg.Validators.NodeDiskType),
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
					//TODO: Move this to image, not plays well with other metadata
					//"startup-script-url": pulumi.String(validator.NodeStartupScript),
					"ssh-keys": pulumi.String(formatSSHKeys(cfg.SSHKeys)),
				},
				// Label the node, so we have unique identifiers for each node
				Labels: pulumi.StringMap{
					"type":  pulumi.String("validator"),
					"index": pulumi.String(fmt.Sprintf("%d", instanceNum)),
				},
				Name:     pulumi.String(name),
				Hostname: pulumi.String(hostname),
				Tags:     pulumi.StringArray{pulumi.String("starnet-validator")}, // Apply firewall rules
				// Set scheduling to preemptible if nodesAutoDelete is true (e.g auto delete after 24h)
				Scheduling: &compute.InstanceSchedulingArgs{
					Preemptible:      pulumi.Bool(cfg.Validators.NodesAutoDelete),
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
				nodes.Validators = append(nodes.Validators, Node{
					Host: hostname,
					IP:   ipStr,
				})
				return ipStr
			})

			instances = append(instances, vm)
			ctx.Export(hostname, ip)
		}

		// Create a firewall rule for all node types
		if err := createFirewall(ctx, cfg, instances); err != nil {
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
		interfaceIPs := make([]interface{}, len(allIPs))
		for i, ip := range allIPs {
			interfaceIPs[i] = ip
		}

		// Wait for all instance/IPs to be available before generating configs
		pulumi.All(interfaceIPs...).ApplyT(func(ips []interface{}) error {
			generateCmd, err := NewCustomCommand(ctx, "generate-configs", func() error {
				return generateNodesConfigs(cfg, nodes)
			}, resources)
			if err != nil {
				ctx.Log.Error(fmt.Sprintf("error generating configs: %v", err), nil)
				return err
			}

			// Sync nodes
			_, err = NewCustomCommand(ctx, "copy-configs", func() error {
				return syncNodes(ctx, nodes, instances)
			}, []pulumi.Resource{generateCmd}) // Wait for generateCmd to finish before copying configs
			if err != nil {
				ctx.Log.Error(fmt.Sprintf("error syncing nodes: %v", err), nil)
				return err
			}
			return nil
		})

		return nil
	})
}
