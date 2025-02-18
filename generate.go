package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Variable for mocking in tests
var execCommand = exec.Command
var execLookPath = exec.LookPath

type Peers []string

const (
	ID_FILE_PATH        = "ids.json"
	CHAIN_STRESSER_PATH = "chain-stresser-deploy"
	VALIDATORS_ID_PATH  = CHAIN_STRESSER_PATH + "/" + "validators" + "/" + ID_FILE_PATH
	SENTRIES_ID_PATH    = CHAIN_STRESSER_PATH + "/" + "sentries" + "/" + ID_FILE_PATH
	INJECTIVE_REPO_PATH = "injective-core"
	// Add path where injectived will be built
	INJECTIVED_BINARY_PATH = INJECTIVE_REPO_PATH + "/build/injectived"
	GOOS                   = "linux"
	GOARCH                 = "amd64"
)

func generateNodesConfigs(cfg Config, nodes Nodes) error {
	// Check if chain-stresser binary exists in PATH
	if err := checkBinaryPath("chain-stresser"); err != nil {
		return err
	}
	// Run binary to generate config files
	cmd := execCommand("chain-stresser", "generate",
		"--accounts-num", fmt.Sprintf("%d", cfg.NetworkConfig.AccountsNum),
		"--validators", fmt.Sprintf("%d", len(nodes.Validators)),
		"--sentries", fmt.Sprintf("%d", len(nodes.Sentries)),
		"--instances", fmt.Sprintf("%d", cfg.NetworkConfig.Instances),
		"--evm", strconv.FormatBool(cfg.NetworkConfig.EVM))

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running chain-stresser: %v", err)
	}

	// Read node IDs from ids.json
	data, err := os.ReadFile(VALIDATORS_ID_PATH)
	if err != nil {
		return fmt.Errorf("error reading ids.json: %v", err)
	}

	var nodeIDs []string
	if err := json.Unmarshal(data, &nodeIDs); err != nil {
		return fmt.Errorf("error parsing ids.json: %v", err)
	}

	// Create a map to store validator index mapping.
	// Validators are not ordered (pulumi creates them in random order, its async action)
	validatorMap := make(map[int]int)
	for i, validator := range nodes.Validators {
		// Extract index from hostname (validator-N)
		var index int
		if _, err := fmt.Sscanf(validator.Host, "starnet-validator-%d", &index); err != nil {
			return fmt.Errorf("failed to parse validator index from hostname %s: %v", validator.Host, err)
		}
		validatorMap[index] = i
	}

	// Assign node IDs based on validator index
	for i := 0; i < len(nodeIDs); i++ {
		if validatorIndex, exists := validatorMap[i]; exists {
			nodes.Validators[validatorIndex].NetworkNodeID = nodeIDs[i]
		} else {
			return fmt.Errorf("validator-%d not found in nodes list", i)
		}
	}

	// build the peer list
	var peers Peers
	for i := 0; i < len(nodes.Validators); i++ {
		peers = append(peers, fmt.Sprintf("%s@%s:26656", nodes.Validators[i].NetworkNodeID, nodes.Validators[i].IP))
	}

	if err := updateValidatorConfigs(peers); err != nil {
		return fmt.Errorf("error updating validator configs: %v", err)
	}

	// Build binary
	if err := prepareInjectived(cfg, nodes); err != nil {
		return fmt.Errorf("error preparing injectived: %v", err)
	}

	return nil
}

func updateValidatorConfigs(peers Peers) error {
	// Convert peers slice to comma-separated string
	peersStr := strings.Join(peers, ",")

	// Loop through each validator directory
	for i := range peers {
		configPath := fmt.Sprintf("%s/%d/config/config.toml", CHAIN_STRESSER_PATH+"/validators", i)

		// Read the existing config file
		content, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("error reading config file %s: %v", configPath, err)
		}

		// Convert content to string and split into lines
		lines := strings.Split(string(content), "\n")

		// Find and replace the persistent_peers line
		found := false
		for i, line := range lines {
			if strings.Contains(line, "persistent_peers") {
				lines[i] = fmt.Sprintf(`persistent_peers = "%s"`, peersStr)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("persistent_peers line not found in config file %s", configPath)
		}

		// Join the lines back together
		newContent := strings.Join(lines, "\n")

		// Write the updated content back to the file
		err = os.WriteFile(configPath, []byte(newContent), 0644)
		if err != nil {
			return fmt.Errorf("error writing config file %s: %v", configPath, err)
		}
	}

	return nil
}

// Pull injective-core from provided branch and build it
func prepareInjectived(cfg Config, nodes Nodes) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %v", err)
	}
	injectiveAbsPath := filepath.Join(workDir, INJECTIVE_REPO_PATH)

	// Clone repository if it doesn't exist
	if _, err := os.Stat(injectiveAbsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(injectiveAbsPath, 0755); err != nil {
			return fmt.Errorf("error creating injective directory: %v", err)
		}

		cmd := execCommand("git", "clone", cfg.InjectiveConfig.Repository, injectiveAbsPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error cloning injective repository: %v", err)
		}
	}

	// Checkout specified branch
	cmd := execCommand("git", "-C", injectiveAbsPath, "checkout", cfg.InjectiveConfig.Branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error checking out branch: %v", err)
	}

	// Build binary
	cmd = execCommand("env", "GOOS="+GOOS, "GOARCH="+GOARCH, "make", "-C", injectiveAbsPath, "install")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building injectived: %v", err)
	}

	// Find binary in GOPATH
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		goPath = filepath.Join(homeDir, "go")
	}

	binaryPath := filepath.Join(goPath, "bin", "injectived")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s after make install", binaryPath)
	}

	// Copy binary to each validator's directory
	for i := range nodes.Validators {
		validatorBinPath := filepath.Join(CHAIN_STRESSER_PATH, "validators", strconv.Itoa(i), "injectived")
		if err := copyFile(binaryPath, validatorBinPath); err != nil {
			return fmt.Errorf("error copying binary to validator %d: %v", i, err)
		}
	}

	// Clean up build directory
	if err := os.RemoveAll(injectiveAbsPath); err != nil {
		return fmt.Errorf("error cleaning up injective directory: %v", err)
	}

	return nil
}

// Helper function to copy files
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	// Close the destination file before changing permissions
	destFile.Close()

	// Set executable permissions (rwxr-xr-x)
	if err := os.Chmod(dst, 0755); err != nil {
		return fmt.Errorf("error setting file permissions: %v", err)
	}

	return nil
}

func checkBinaryPath(binary string) error {
	_, err := execLookPath(binary)
	if err != nil {
		return fmt.Errorf("binary %s not found in PATH: %v", binary, err)
	}

	return nil
}
