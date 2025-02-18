package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to check if chain-stresser binary exists
func checkChainStresserBinary(t *testing.T) {
	_, err := exec.LookPath("chain-stresser")
	if err != nil {
		t.Fatal("chain-stresser binary not found in PATH. Please build it first.")
	}
}

func TestAChainStresserBinaryExists(t *testing.T) {
	_, err := exec.LookPath("chain-stresser")
	if err != nil {
		t.Fatal("chain-stresser binary not found in PATH. Please build it first.")
	}
}

func TestGenerateNodesConfigs(t *testing.T) {
	cfg := Config{
		NetworkConfig: CometConfig{
			AccountsNum: 10,
			Validators:  2,
			Sentries:    2,
			Instances:   1,
			EVM:         false,
		},
		InjectiveConfig: InjectiveConfig{
			Repository: "https://github.com/InjectiveLabs/injective-core",
			Branch:     "master",
		},
	}

	nodes := Nodes{
		Validators: []Node{
			{Host: "starnet-validator-0", IP: "10.0.0.1"},
			{Host: "starnet-validator-1", IP: "10.0.0.2"},
		},
		Sentries: []Node{
			{Host: "sentry0", IP: "10.0.1.1"},
			{Host: "sentry1", IP: "10.0.1.2"},
		},
	}

	err := generateNodesConfigs(cfg, nodes)
	if err != nil {
		t.Fatalf("generateNodesConfigs failed: %v", err)
	}

	// Verify the configs were created
	for i := range nodes.Validators {
		configPath := fmt.Sprintf("%s/validators/%d/config/config.toml", CHAIN_STRESSER_PATH, i)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config not found at expected path: %s", configPath)
		}
	}

	for i := range nodes.Sentries {
		configPath := fmt.Sprintf("%s/sentry-nodes/%d/config/config.toml", CHAIN_STRESSER_PATH, i)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config not found at expected path: %s", configPath)
		}
	}
}

func TestZUpdateValidatorConfigs(t *testing.T) {
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
	err = updateValidatorConfigs(peers)
	if err != nil {
		t.Fatalf("updateValidatorConfigs failed: %v", err)
	}

	// Verify the results
	expectedPeers := `node1@192.168.1.1:26656,node2@192.168.1.2:26656`
	for i := 0; i < 2; i++ {
		configPath := filepath.Join(projectRoot, CHAIN_STRESSER_PATH, fmt.Sprintf("validators/%d/config/config.toml", i))
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated config file: %v", err)
		}

		// Check if the persistent_peers line was updated correctly
		lines := strings.Split(string(content), "\n")
		found := false
		for _, line := range lines {
			if strings.Contains(line, "persistent_peers") {
				if !strings.Contains(line, expectedPeers) {
					t.Errorf("Incorrect persistent_peers value.\nExpected to contain: %s\nGot: %s", expectedPeers, line)
				}
				found = true
				break
			}
		}
		if !found {
			t.Error("persistent_peers line not found in config file")
		}
	}
}

// No mocking here, this test actually clone the repo and build the binary from provided branch.
func TestPrepareInjectived(t *testing.T) {
	cfg := Config{
		NetworkConfig: CometConfig{
			AccountsNum: 10,
			Validators:  2,
			Sentries:    2,
			Instances:   1,
			EVM:         false,
		},
		InjectiveConfig: InjectiveConfig{
			Repository: "https://github.com/InjectiveLabs/injective-core",
			Branch:     "master",
		},
	}

	nodes := Nodes{
		Validators: []Node{
			{Host: "starnet-validator-0", IP: "10.0.0.1"},
			{Host: "starnet-validator-1", IP: "10.0.0.2"},
		},
		Sentries: []Node{
			{Host: "sentry0", IP: "10.0.1.1"},
			{Host: "sentry1", IP: "10.0.1.2"},
		},
	}

	err := prepareInjectived(cfg, nodes)
	if err != nil {
		t.Fatalf("prepareInjectived failed: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "copy-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file with some content
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	dstPath := filepath.Join(tmpDir, "destination.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify the content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination content does not match source. Got %s, want %s", dstContent, content)
	}
}
