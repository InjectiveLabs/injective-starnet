package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Network struct {
	Name string `json:"networkName"`
}

type NodePool struct {
	NodePoolName       string   `json:"nodePoolName"`
	NodePoolSize       int      `json:"nodePoolSize"`
	NodeImage          string   `json:"nodeImage"`
	NodeStartupScript  string   `json:"nodeStartupScript"`
	NodeMachineType    string   `json:"nodeMachineType"`
	NodeDiskSizeGB     int      `json:"nodeDiskSizeGB"`
	NodeDiskType       string   `json:"nodeDiskType"`
	NodesAutoDelete    bool     `json:"nodesAutoDelete"`
	NodePoolLabel      []string `json:"nodePoolLabel"`
	NodeRegions        []string `json:"nodeRegions"`
	NodeZonesPerRegion int      `json:"nodeZonesPerRegion"`
	NodePorts          []string `json:"nodePorts,omitempty"`
}

type CometConfig struct {
	AccountsNum int  `json:"accountsNum"`
	Validators  int  `json:"validators"`
	Sentries    int  `json:"sentries"`
	Instances   int  `json:"instances"`
	EVM         bool `json:"evm"`
}

// Add new SSH key structure
type SSHKey struct {
	Username string `json:"username"`
	Key      string `json:"key"`
}

type SSHKeys struct {
	Keys []SSHKey
}

type InjectiveConfig struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
}

type Config struct {
	Project         string      `json:"project"`
	SSHKeys         SSHKeys     `json:"sshKeys"`
	Validators      NodePool    `json:"validators"`
	Sentries        NodePool    `json:"sentries"`
	NetworkConfig   CometConfig `json:"comet"`
	InjectiveConfig InjectiveConfig
}

func loadConfig(ctx *pulumi.Context) (Config, error) {
	gcpConfig := config.New(ctx, "gcp")
	starnetConfig := config.New(ctx, "starnet")

	project := gcpConfig.Require("project")

	var sshKeys SSHKeys
	starnetConfig.RequireObject("sshKeys", &sshKeys)

	var injectiveConfig InjectiveConfig
	starnetConfig.RequireObject("injective", &injectiveConfig)

	var nodePools []NodePool
	starnetConfig.RequireObject("nodePools", &nodePools)
	validators := nodePools[0]
	sentries := nodePools[1]

	var cometConfig CometConfig
	starnetConfig.RequireObject("comet", &cometConfig)

	return Config{
		Project:         project,
		SSHKeys:         sshKeys,
		Validators:      validators,
		Sentries:        sentries,
		NetworkConfig:   cometConfig,
		InjectiveConfig: injectiveConfig,
	}, nil
}
