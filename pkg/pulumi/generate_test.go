package pulumi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAChainStresserBinaryExists(t *testing.T) {
	_, err := exec.LookPath("chain-stresser")
	if err != nil {
		t.Fatal("chain-stresser binary not found in PATH. Please build it first.")
	}
}

func TestGenerateNodesConfigs(t *testing.T) {
	cfg := Config{
		InjectiveConfig: InjectiveConfig{
			Repository: "https://github.com/InjectiveLabs/injective-core",
			Branch:     "master",
		},
	}

	nodes := Nodes{
		Validators: []Node{
			{Host: "starnet-validators-0.injective.network", IP: "10.0.0.1"},
			{Host: "starnet-validators-1.injective.network", IP: "10.0.0.2"},
		},
		Sentries: []Node{
			{Host: "starnet-sentry-nodes-0.injective.network", IP: "10.0.1.1"},
		},
	}

	// Test for validators
	err := GenerateNodesConfigs(cfg, nodes, VALIDATORS_TYPE)
	if err != nil {
		t.Fatalf("GenerateNodesConfigs failed for validators: %v", err)
	}

	// Verify the configs were created for validators
	for i := range nodes.Validators {
		configPath := fmt.Sprintf("%s/validators/%d/config/config.toml", CHAIN_STRESSER_PATH, i)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config not found at expected path for validator: %s", configPath)
		}
	}

	// Test for sentries
	err = GenerateNodesConfigs(cfg, nodes, SENTRIES_TYPE)
	if err != nil {
		t.Fatalf("GenerateNodesConfigs failed for sentries: %v", err)
	}

	// Verify the configs were created for sentries
	for i := range nodes.Sentries {
		configPath := fmt.Sprintf("%s/sentry-nodes/%d/config/config.toml", CHAIN_STRESSER_PATH, i)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config not found at expected path for sentry: %s", configPath)
		}
	}
}

func TestZUpdateNodesConfigs(t *testing.T) {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Check if the config.toml exists from the actual chain-stresser generation
	configPath := filepath.Join(projectRoot, CHAIN_STRESSER_PATH, "validators/0/config/config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Skipping test as required files don't exist. Please run chain-stresser generate first.")
	}

	// Test peers
	peers := Peers{
		"node1@192.168.1.1:26656",
		"node2@192.168.1.2:26656",
	}

	// Run the function
	nodes := make([]Node, 2)
	nodes[0] = Node{Host: "starnet-validators-0.injective.network", IP: "10.0.0.1"}
	nodes[1] = Node{Host: "starnet-validators-1.injective.network", IP: "10.0.0.2"}

	err = updateNodesConfigs(peers, nodes, VALIDATORS_TYPE)
	if err != nil {
		t.Fatalf("updateNodesConfigs failed: %v", err)
	}

	// Test for sentry nodes
	sentryNodes := make([]Node, 2)
	sentryNodes[0] = Node{Host: "starnet-sentry-nodes-0.injective.network", IP: "10.0.1.1"}
	sentryNodes[1] = Node{Host: "starnet-sentry-nodes-1.injective.network", IP: "10.0.1.2"}

	err = updateNodesConfigs(peers, sentryNodes, SENTRIES_TYPE)
	if err != nil {
		t.Fatalf("updateNodesConfigs failed for sentries: %v", err)
	}

	// Verify the results for both validators and sentries
	expectedPeers := `node1@192.168.1.1:26656,node2@192.168.1.2:26656`

	// Check validators
	for i := 0; i < 2; i++ {
		configPath := filepath.Join(projectRoot, CHAIN_STRESSER_PATH, fmt.Sprintf("validators/%d/config/config.toml", i))
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated validator config file: %v", err)
		}

		// Check if the persistent_peers line was updated correctly
		lines := strings.Split(string(content), "\n")
		found := false
		for _, line := range lines {
			if strings.Contains(line, "persistent_peers") {
				if !strings.Contains(line, expectedPeers) {
					t.Errorf("Incorrect persistent_peers value in validator config.\nExpected to contain: %s\nGot: %s", expectedPeers, line)
				}
				found = true
				break
			}
		}
		if !found {
			t.Error("persistent_peers line not found in validator config file")
		}
	}

	// Check sentry nodes
	for i := 0; i < 2; i++ {
		configPath := filepath.Join(projectRoot, CHAIN_STRESSER_PATH, fmt.Sprintf("sentry-nodes/%d/config/config.toml", i))
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated sentry config file: %v", err)
		}

		// Check if the persistent_peers line was updated correctly
		lines := strings.Split(string(content), "\n")
		found := false
		for _, line := range lines {
			if strings.Contains(line, "persistent_peers") {
				if !strings.Contains(line, expectedPeers) {
					t.Errorf("Incorrect persistent_peers value in sentry config.\nExpected to contain: %s\nGot: %s", expectedPeers, line)
				}
				found = true
				break
			}
		}
		if !found {
			t.Error("persistent_peers line not found in sentry config file")
		}
	}
}
