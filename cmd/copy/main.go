package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Server struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	ID       string `json:"id"`
}

const (
	jsonFile     = "storage.json"
	localFolder  = ".injectived/"
	remoteUser   = "injectived"
	remoteFolder = "/home/injectived/.injectived"
	cmdStop      = "sudo systemctl stop injectived"
	cmdStart     = "sudo systemctl start injectived"
)

func main() {
	// Read and parse JSON file
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("Failed to read JSON file: %v\n", err)
		os.Exit(1)
	}

	var servers []Server
	if err := json.Unmarshal(data, &servers); err != nil {
		fmt.Printf("Failed to parse %s: %v\n", jsonFile, err)
		os.Exit(1)
	}

	// Check if local state folder exists
	if _, err := os.Stat(localFolder); os.IsNotExist(err) {
		fmt.Printf("Folder does NOT exist: %s, no state to copy\n", localFolder)
		os.Exit(1)
	}
	// Ensure rsync is installed before proceeding
	checkRsync()

	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			executeSync(ip)
		}(server.IP)
	}

	wg.Wait()
}

func executeSync(ip string) {

	if err := sshCommand(ip, cmdStop); err != nil {
		fmt.Printf("Failed to execute %s on %s: %v\n", cmdStop, ip, err)
		return
	}

	rsyncCmd := exec.Command("rsync", "-avz", "--progress", "-e", "ssh -o StrictHostKeyChecking=no", localFolder, fmt.Sprintf("%s@%s:%s", remoteUser, ip, remoteFolder))

	// Stream live output
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		fmt.Printf("Sync failed for %s: %v\n", ip, err)
		return
	}

	fmt.Printf("✅ Sync completed for %s\n", ip)
	fmt.Printf("🚀 Starting injectived on %s...\n", ip)
	if err := sshCommand(ip, cmdStart); err != nil {
		fmt.Printf("Failed to execute %s on %s: %v\n", cmdStart, ip, err)
		return
	}
}

func sshCommand(ip, command string) error {
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", remoteUser, ip), command)
	return cmd.Run()
}

func checkRsync() {
	cmd := exec.Command("which", "rsync")
	if err := cmd.Run(); err != nil {
		fmt.Println("rsync is not installed. Please install it and try again.")
		os.Exit(1)
	}
}
