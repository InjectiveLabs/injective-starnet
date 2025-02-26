package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/InjectiveLabs/injective-starnet/storage"
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
)

func GenerateNodesConfigs(cfg Config, nodes Nodes) error {
	store := storage.NewFileStore("./storage.json")

	if err := CheckArtifacts(CHAIN_STRESSER_PATH, nodes); err != nil {
		return fmt.Errorf("error checking artifacts: %v", err)
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
	validatorMap := make(map[int]int)
	maxIndex := -1

	// Sort validators by their index number
	sort.Slice(nodes.Validators, func(i, j int) bool {
		var index1, index2 int
		fmt.Sscanf(nodes.Validators[i].Host, "starnet-validator-%d", &index1)
		fmt.Sscanf(nodes.Validators[j].Host, "starnet-validator-%d", &index2)
		return index1 < index2
	})

	// Create mapping with sorted validators
	for i, validator := range nodes.Validators {
		var index int
		if _, err := fmt.Sscanf(validator.Host, "starnet-validator-%d", &index); err != nil {
			return fmt.Errorf("failed to parse validator index from hostname %s: %v", validator.Host, err)
		}
		validatorMap[index] = i
		if index > maxIndex {
			maxIndex = index
		}
	}

	// Verify we have all validators from 0 to maxIndex
	for i := 0; i <= maxIndex; i++ {
		if _, exists := validatorMap[i]; !exists {
			return fmt.Errorf("missing validator-%d in sequence (have %d validators)", i, len(nodes.Validators))
		}
	}

	// Verify nodeIDs count matches validators count
	if len(nodeIDs) != len(nodes.Validators) {
		return fmt.Errorf("mismatch between nodeIDs (%d) and validators (%d)", len(nodeIDs), len(nodes.Validators))
	}

	// Assign nodeIDs in order
	for i := range nodes.Validators {
		nodes.Validators[i].NetworkNodeID = nodeIDs[i]
	}

	// Build the peer list in order
	var peers Peers
	for i := range nodes.Validators {
		peers = append(peers, fmt.Sprintf("%s@%s:26656",
			nodes.Validators[i].NetworkNodeID,
			nodes.Validators[i].IP))
	}

	if err := updateValidatorConfigs(peers); err != nil {
		return fmt.Errorf("error updating validator configs: %v", err)
	}

	// Store the records in the storage
	var records []storage.Record
	for i := range nodes.Validators {
		records = append(records, storage.Record{
			Hostname: nodes.Validators[i].Host,
			IP:       nodes.Validators[i].IP,
			ID:       nodes.Validators[i].NetworkNodeID,
		})
	}
	store.SetAll(records)

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

// Check if all artifacts are present
func CheckArtifacts(path string, nodes Nodes) error {
	// Check if if generated for validators
	validatorsPath := path + "/validators"
	if _, err := os.Stat(validatorsPath); os.IsNotExist(err) {
		return fmt.Errorf("validators directory not found in %s", path)
	}

	// Loop over validators and check if injectived binary and libwasmvm.x86_64.so are present
	for i := range nodes.Validators {
		validatorDir := fmt.Sprintf("%s/%d", validatorsPath, i)

		// Check if injectived binary is present and is executable
		injectivedPath := validatorDir + "/injectived"
		if _, err := os.Stat(injectivedPath); os.IsNotExist(err) {
			return fmt.Errorf("injectived binary not found in %s", validatorDir)
		}

		// Check if libwasmvm.x86_64.so is present
		wasmvmPath := validatorDir + "/libwasmvm.x86_64.so"
		if _, err := os.Stat(wasmvmPath); os.IsNotExist(err) {
			return fmt.Errorf("libwasmvm.x86_64.so not found in %s", validatorDir)
		}
	}

	return nil
}
