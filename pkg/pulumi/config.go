package pulumi

import (
	"embed"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gopkg.in/yaml.v3"
)

//go:embed Pulumi.yaml Pulumi.starnet.yaml
var configFiles embed.FS

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
	Project         string   `json:"project"`
	SSHKeys         SSHKeys  `json:"sshKeys"`
	Validators      NodePool `json:"validators"`
	Sentries        NodePool `json:"sentries"`
	InjectiveConfig InjectiveConfig
}

type PulumiConfig struct {
	Config map[string]interface{} `yaml:"config"`
}

// GetPulumiYAML returns the embedded Pulumi.yaml content
func GetPulumiYAML() ([]byte, error) {
	return configFiles.ReadFile("Pulumi.yaml")
}

// LoadEmbeddedConfig loads and parses the embedded Pulumi.starnet.yaml
func LoadEmbeddedConfig() (map[string]interface{}, error) {
	content, err := configFiles.ReadFile("Pulumi.starnet.yaml")
	if err != nil {
		return nil, err
	}

	var config PulumiConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return config.Config, nil
}

// LoadConfig loads the configuration from Pulumi config
func LoadConfig(ctx *pulumi.Context) (Config, error) {
	gcpConfig := config.New(ctx, "gcp")
	cfg := config.New(ctx, "")

	project := gcpConfig.Require("project")

	var sshKeys SSHKeys
	cfg.RequireObject("sshKeys", &sshKeys)

	var injectiveConfig InjectiveConfig
	cfg.RequireObject("injective", &injectiveConfig)

	var nodePools []NodePool
	cfg.RequireObject("nodePools", &nodePools)
	validators := nodePools[0]
	sentries := nodePools[1]

	return Config{
		Project:         project,
		SSHKeys:         sshKeys,
		Validators:      validators,
		Sentries:        sentries,
		InjectiveConfig: injectiveConfig,
	}, nil
}
