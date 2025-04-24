# Remote Command Executor

A Go package for executing commands on remote machines in parallel using SSH with agent forwarding.

## Features

- Execute commands on multiple remote hosts in parallel
- SSH agent forwarding for secure authentication
- Capture stdout, stderr, and exit codes
- Simple and easy to use API

## Requirements

- Go 1.21 or higher
- SSH agent running on your local machine
- SSH access to remote hosts

## Installation

```bash
go get github.com/nmarcetic/injective-starnet/pkg/exec
```

## Usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nmarcetic/injective-starnet/pkg/exec"
)

func main() {
	// Check if SSH_AUTH_SOCK is set
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		log.Fatal("SSH_AUTH_SOCK environment variable is not set. Please ensure SSH agent is running.")
	}

	// Create an executor
	executor := &exec.Executor{
		Hosts: []string{"host1.example.com", "host2.example.com"},
		Cmd:   "git pull",
	}

	// Execute the command on all hosts
	results := executor.Execute()

	// Print the results
	for _, result := range results {
		fmt.Printf("Host: %s\n", result.Host)
		if result.Err != nil {
			fmt.Printf("Error: %v\n", result.Err)
		} else {
			fmt.Printf("Exit Code: %d\n", result.ExitCode)
			if result.Stdout != "" {
				fmt.Printf("Stdout:\n%s\n", result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Printf("Stderr:\n%s\n", result.Stderr)
			}
		}
		fmt.Println("-----------------------------------")
	}
}
```

## SSH Agent Forwarding

This executor uses SSH agent forwarding to securely use your local SSH keys on remote machines. This is particularly useful for operations like git clone, git pull, etc., where you need to authenticate with remote repositories.

To use SSH agent forwarding:

1. Make sure your SSH agent is running on your local machine
2. Add your SSH keys to the agent: `ssh-add ~/.ssh/id_rsa`
3. Ensure the `SSH_AUTH_SOCK` environment variable is set

## License

MIT

# :star2: Starnet is galactic-scale orchestrator for distributed cosmos network.

  

## What is Starnet?

  

Starnet is a powerful and scalable testnet deployment system.

  

## How does it work?

  

Starnet uses Pulumi to create and manage network of nodes.

TBA

  

## Requirements

TBA

  

## How do I use it?

  

### Use Github Actions to deploy Starnet

Currently only [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job works, you can use it nuke the running network.

  

### Use Local Machine to deploy Starnet

  

Currently we have set up a local machine to deploy Starnet on mainne-devlab node (playground machine).

  

ssh config

  

    Host mainne-devlab
    
      
    
     HostName 37.59.23.104
    
      
    
     User root
    
      
    
     Port 22
    
      
    
    I dentityFile ~/.ssh/id_rsa

  
  

When you ssh, run the following commands:

  

    cd injective/injective-starnet
    git pull  
    gcloud auth login --cred-file=gcloud.json
    export GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/gcloud.json

  
  

#### Tweak local script that wraps chain-stresser, adjust the number of validators
Note: For testing purpose use smaller network 2-10 validators max.

    vi scripts/local.sh
    # generate artifacts
    ./scripts/local.sh # This should output you chain-stresser-deploy folder in same path

#### Align validators with the number of VM's we need to create

    vi Pulumi.starnet.yaml # change the nodePoolSize to fit the value you set in scripts/local.sh
    
#### Spin up the network

    pulumi up -y

#### When pulumi finishes, you will get validators IP's in output

Copy any validator IP and  ssh (from that devlab machine, but you should be able to ssh from your machine also)

    ssh injectived@<validator_ip_from_output>
    
Check the network status with `gex` 


#### To nuke the network

    pulumi destroy -y

**NOTE:** if you are not sure if you nuked network properly , go to injective-starnet repo and manually run [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job (top right corner have "run the workflow" buttin").
All ports are opened to public, p2p, RPC , gRPC 
  
  
  
