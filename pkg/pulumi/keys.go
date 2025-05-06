package pulumi

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetStarnetKey returns the path to the starnet_key file
// If the file doesn't exist in the current directory, it looks for it in the keys directory
func GetStarnetKey() (string, error) {
	// First check if the file exists in the current directory
	keyPath := "starnet_key"
	if _, err := os.Stat(keyPath); err == nil {
		return keyPath, nil
	}

	// Then check in the keys directory
	keysPath := filepath.Join("keys", "starnet_key")
	if _, err := os.Stat(keysPath); err == nil {
		return keysPath, nil
	}

	// Finally check in the pkg/pulumi/keys directory
	pkgKeysPath := filepath.Join("pkg", "pulumi", "keys", "starnet_key")
	if _, err := os.Stat(pkgKeysPath); err == nil {
		return pkgKeysPath, nil
	}

	return "", fmt.Errorf("starnet_key not found in current directory, keys directory, or pkg/pulumi/keys directory")
}
