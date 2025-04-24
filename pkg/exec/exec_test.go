package exec

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestExecutor_Execute tests the Execute method with real IPs
func TestExecutor_Execute(t *testing.T) {
	// Skip if SSH_AUTH_SOCK is not set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Skip("SSH_AUTH_SOCK environment variable is not set. Skipping test.")
	}

	// Get test hosts from environment variable or use default
	hosts := getTestHosts(t)
	if len(hosts) == 0 {
		t.Skip("No test hosts provided. Skipping test.")
	}

	// Get username from environment variable or use default
	username := getTestUsername(t)

	// Create an executor
	executor := &Executor{
		Hosts:    hosts,
		Cmd:      "echo 'Hello from remote host'",
		Username: username,
	}

	// Execute the command on all hosts
	results := executor.Execute()

	// Check the results
	if len(results) != len(hosts) {
		t.Errorf("Expected %d results, got %d", len(hosts), len(results))
	}

	for i, result := range results {
		if result.Err != nil {
			t.Errorf("Host %s: Error executing command: %v", hosts[i], result.Err)
		} else {
			if result.ExitCode != 0 {
				t.Errorf("Host %s: Expected exit code 0, got %d", hosts[i], result.ExitCode)
			}
			if result.Stdout == "" {
				t.Errorf("Host %s: Expected non-empty stdout", hosts[i])
			}
		}
	}
}

// TestExecutor_ExecuteWithGit tests the Execute method with git commands
func TestExecutor_ExecuteWithGit(t *testing.T) {
	// Skip if SSH_AUTH_SOCK is not set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Skip("SSH_AUTH_SOCK environment variable is not set. Skipping test.")
	}

	// Get test hosts from environment variable or use default
	hosts := getTestHosts(t)
	if len(hosts) == 0 {
		t.Skip("No test hosts provided. Skipping test.")
	}

	// Get username from environment variable or use default
	username := getTestUsername(t)

	// Create an executor
	executor := &Executor{
		Hosts:    hosts,
		Cmd:      "git --version",
		Username: username,
	}

	// Execute the command on all hosts
	results := executor.Execute()

	// Check the results
	if len(results) != len(hosts) {
		t.Errorf("Expected %d results, got %d", len(hosts), len(results))
	}

	for i, result := range results {
		if result.Err != nil {
			t.Errorf("Host %s: Error executing command: %v", hosts[i], result.Err)
		} else {
			if result.ExitCode != 0 {
				t.Errorf("Host %s: Expected exit code 0, got %d", hosts[i], result.ExitCode)
			}
			if result.Stdout == "" {
				t.Errorf("Host %s: Expected non-empty stdout", hosts[i])
			}
			// Check if the output contains "git version"
			if !contains(result.Stdout, "git version") {
				t.Errorf("Host %s: Expected stdout to contain 'git version', got: %s", hosts[i], result.Stdout)
			}
		}
	}
}

// TestExecutor_ExecuteWithError tests the Execute method with a command that should fail
func TestExecutor_ExecuteWithError(t *testing.T) {
	// Skip if SSH_AUTH_SOCK is not set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Skip("SSH_AUTH_SOCK environment variable is not set. Skipping test.")
	}

	// Get test hosts from environment variable or use default
	hosts := getTestHosts(t)
	if len(hosts) == 0 {
		t.Skip("No test hosts provided. Skipping test.")
	}

	// Get username from environment variable or use default
	username := getTestUsername(t)

	// Create an executor with a command that should fail
	executor := &Executor{
		Hosts:    hosts,
		Cmd:      "nonexistentcommand",
		Username: username,
	}

	// Execute the command on all hosts
	results := executor.Execute()

	// Check the results
	if len(results) != len(hosts) {
		t.Errorf("Expected %d results, got %d", len(hosts), len(results))
	}

	for i, result := range results {
		if result.Err != nil {
			t.Errorf("Host %s: Expected error in result.Err, got error in result: %v", hosts[i], result.Err)
		}
		if result.ExitCode == 0 {
			t.Errorf("Host %s: Expected non-zero exit code, got 0", hosts[i])
		}
	}
}

// TestExecutor_ExecuteWithTimeout tests the Execute method with a timeout
func TestExecutor_ExecuteWithTimeout(t *testing.T) {
	// Skip if SSH_AUTH_SOCK is not set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Skip("SSH_AUTH_SOCK environment variable is not set. Skipping test.")
	}

	// Get test hosts from environment variable or use default
	hosts := getTestHosts(t)
	if len(hosts) == 0 {
		t.Skip("No test hosts provided. Skipping test.")
	}

	// Get username from environment variable or use default
	username := getTestUsername(t)

	// Create an executor with a command that should timeout
	executor := &Executor{
		Hosts:    hosts,
		Cmd:      "sleep 10",
		Username: username,
	}

	// Execute the command on all hosts with a timeout
	done := make(chan bool)
	var results []Result

	go func() {
		results = executor.Execute()
		done <- true
	}()

	// Wait for the command to complete or timeout
	select {
	case <-done:
		// Command completed
	case <-time.After(5 * time.Second):
		t.Fatal("Command timed out after 5 seconds")
	}

	// Check the results
	if len(results) != len(hosts) {
		t.Errorf("Expected %d results, got %d", len(hosts), len(results))
	}
}

// getTestHosts returns the test hosts from the environment variable or default to localhost
func getTestHosts(t *testing.T) []string {
	// Get test hosts from environment variable
	testHosts := os.Getenv("TEST_EXEC_HOSTS")
	if testHosts != "" {
		return strings.Split(testHosts, ",")
	}

	// Check if we can connect to localhost:22
	conn, err := net.DialTimeout("tcp", "localhost:22", 2*time.Second)
	if err != nil {
		t.Skip("Could not connect to localhost:22. Skipping test.")
		return nil
	}
	conn.Close()

	return []string{"localhost"}
}

// getTestUsername returns the test username from the environment variable or default to current user
func getTestUsername(t *testing.T) string {
	// Get test username from environment variable
	username := os.Getenv("TEST_EXEC_USERNAME")
	if username != "" {
		return username
	}

	// Default to current user
	currentUser, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine current user. Skipping test.")
		return ""
	}

	// Extract username from home directory path
	parts := strings.Split(currentUser, "/")
	if len(parts) < 2 {
		t.Skip("Could not determine current user. Skipping test.")
		return ""
	}

	return parts[len(parts)-1]
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && s[0:len(substr)] == substr
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Check if we're running in a CI environment
	if os.Getenv("CI") != "" {
		// In CI, we might not have SSH agent running
		// Try to start one if needed
		if os.Getenv("SSH_AUTH_SOCK") == "" {
			// Try to start ssh-agent
			cmd := exec.Command("ssh-agent")
			output, err := cmd.Output()
			if err == nil {
				// Parse the output to get the SSH_AUTH_SOCK
				// This is a simple implementation and might not work in all environments
				fmt.Printf("Started ssh-agent: %s\n", output)
				// In a real implementation, you would parse the output and set SSH_AUTH_SOCK
			}
		}
	}

	// Run the tests
	os.Exit(m.Run())
}
