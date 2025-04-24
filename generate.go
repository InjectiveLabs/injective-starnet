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
	SENTRIES_ID_PATH    = CHAIN_STRESSER_PATH + "/" + "sentry-nodes" + "/" + ID_FILE_PATH
	INJECTIVE_REPO_PATH = "injective-core"
	// Add path where injectived will be built
	INJECTIVED_BINARY_PATH = INJECTIVE_REPO_PATH + "/build/injectived"
	VALIDATORS_TYPE        = "validators"
	SENTRIES_TYPE          = "sentry-nodes"
	DEFAULT_TYPE           = VALIDATORS_TYPE
)

func GenerateNodesConfigs(cfg Config, nodes Nodes, nodeType string) error {
	if nodeType == "" || nodeType != VALIDATORS_TYPE && nodeType != SENTRIES_TYPE {
		nodeType = DEFAULT_TYPE
	}

	store := storage.NewFileStore("./storage.json")

	// Read node IDs from ids.json
	data, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", CHAIN_STRESSER_PATH, nodeType, ID_FILE_PATH))
	if err != nil {
		return fmt.Errorf("error reading %s ids.json: %v", nodeType, err)
	}
	var nodeIDs []string
	if err := json.Unmarshal(data, &nodeIDs); err != nil {
		return fmt.Errorf("error parsing ids.json for %s: %v", nodeType, err)
	}

	// Determine which slice to work with
	var nodeSlice []Node
	switch nodeType {
	case VALIDATORS_TYPE:
		nodeSlice = nodes.Validators
	case SENTRIES_TYPE:
		nodeSlice = nodes.Sentries
	default:
		return fmt.Errorf("unknown node type: %s", nodeType)
	}

	// Create a map to store index mapping.
	nodeMap := make(map[int]int)
	maxIndex := -1

	// Sort nodes by their index number
	sort.Slice(nodeSlice, func(i, j int) bool {
		var index1, index2 int
		fmt.Sscanf(nodeSlice[i].Host, fmt.Sprintf("starnet-%s-%%d.injective.network", nodeType), &index1)
		fmt.Sscanf(nodeSlice[j].Host, fmt.Sprintf("starnet-%s-%%d.injective.network", nodeType), &index2)
		return index1 < index2
	})

	// Create mapping with sorted nodes
	for i, node := range nodeSlice {
		var index int
		if _, err := fmt.Sscanf(node.Host, fmt.Sprintf("starnet-%s-%%d.injective.network", nodeType), &index); err != nil {
			return fmt.Errorf("failed to parse index from hostname %s: %v", node.Host, err)
		}
		nodeMap[index] = i
		if index > maxIndex {
			maxIndex = index
		}
	}

	// Verify we have all nodes from 0 to maxIndex
	for i := 0; i <= maxIndex; i++ {
		if _, exists := nodeMap[i]; !exists {
			return fmt.Errorf("missing %s-%d in sequence (have %d %s)", nodeType, i, len(nodeSlice), nodeType)
		}
	}

	// Verify nodeIDs count matches nodes count
	if len(nodeIDs) != len(nodeSlice) {
		return fmt.Errorf("mismatch between nodeIDs (%d) and nodes (%d) for %s. Node IDs: %v, Nodes: %v",
			len(nodeIDs), len(nodeSlice), nodeType,
			nodeIDs,
			func() []string {
				result := make([]string, len(nodeSlice))
				for i, node := range nodeSlice {
					result[i] = node.Host
				}
				return result
			}())
	}

	// Assign nodeIDs in order
	for i := range nodeSlice {
		nodeSlice[i].NetworkNodeID = nodeIDs[i]
	}
	// Build the peer list in order
	var peers Peers
	if nodeType == VALIDATORS_TYPE {
		var records []storage.Record
		for i := range nodes.Validators {
			peers = append(peers, fmt.Sprintf("%s@%s:26656",
				nodes.Validators[i].NetworkNodeID,
				nodes.Validators[i].IP))
			records = append(records, storage.Record{
				Hostname: nodeSlice[i].Host,
				IP:       nodeSlice[i].IP,
				ID:       nodeSlice[i].NetworkNodeID,
			})
		}
		store.SetAll(records)
	} else if nodeType == SENTRIES_TYPE {
		// For sentry nodes, get validator peers from storage
		records, err := store.GetAll()
		if err != nil {
			return fmt.Errorf("error reading validator records from storage: %v", err)
		}

		// Build peers list from storage records
		for _, record := range records {
			peers = append(peers, fmt.Sprintf("%s@%s:26656",
				record.ID,
				record.IP))
		}
	}

	if err := updateNodesConfigs(peers, nodeSlice, nodeType); err != nil {
		return fmt.Errorf("error updating node configs: %v", err)
	}

	return nil
}

func updateNodesConfigs(peers Peers, nodeSlice []Node, nodeType string) error {
	// Convert peers slice to comma-separated string
	peersStr := strings.Join(peers, ",")

	// Loop through each validator directory
	for i := range nodeSlice {
		configPath := fmt.Sprintf("%s/%d/config/config.toml", CHAIN_STRESSER_PATH+"/"+nodeType, i)

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

	// Check if if generated for sentry nodes
	sentryNodesPath := path + "/sentry-nodes"
	if _, err := os.Stat(sentryNodesPath); os.IsNotExist(err) {
		return fmt.Errorf("sentry nodes directory not found in %s", path)
	}

	return nil
}
