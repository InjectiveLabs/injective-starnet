package pulumi

import (
	"embed"
	"fmt"
	"os"
	"sync"
)

//go:embed keys/starnet_key
var embeddedKeys embed.FS

var (
	tmpKeyPath string
	tmpKeyOnce sync.Once
)

// GetStarnetKey returns the path to a temporary file containing the starnet_key
func GetStarnetKey() (string, error) {
	var err error
	tmpKeyOnce.Do(func() {
		keyData, readErr := embeddedKeys.ReadFile("keys/starnet_key")
		if readErr != nil {
			err = fmt.Errorf("failed to read embedded starnet_key: %w", readErr)
			return
		}

		tmpFile, createErr := os.CreateTemp("", "starnet_key_*")
		if createErr != nil {
			err = fmt.Errorf("failed to create temporary file: %w", createErr)
			return
		}

		if _, writeErr := tmpFile.Write(keyData); writeErr != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			err = fmt.Errorf("failed to write key to temporary file: %w", writeErr)
			return
		}

		if closeErr := tmpFile.Close(); closeErr != nil {
			os.Remove(tmpFile.Name())
			err = fmt.Errorf("failed to close temporary file: %w", closeErr)
			return
		}

		tmpKeyPath = tmpFile.Name()
	})

	if err != nil {
		return "", err
	}

	return tmpKeyPath, nil
}

// CleanupStarnetKey removes the temporary key file
func CleanupStarnetKey() error {
	if tmpKeyPath != "" {
		return os.Remove(tmpKeyPath)
	}
	return nil
}
