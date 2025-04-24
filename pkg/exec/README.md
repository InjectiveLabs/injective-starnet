# Executor Package

Simple [Sup inspired](https://github.com/pressly/sup) Go package for executing commands on remote machines via SSH with local agent forwarding for secure key usage and parallel execution.

## Features

- Execute commands on multiple remote hosts in parallel
- SSH agent forwarding for secure authentication (your private keys never leave your local machine Instead, when the remote server challenges your SSH client for authentication, your local SSH agent signs the challenge, and only the signature is sent to the remote server).
- Configurable username for SSH connections
- Detailed execution results including stdout, stderr, and exit codes

## Usage

```go
import "github.com/InjectiveLabs/injective-starnet/pkg/exec"

// Create an executor
executor := &exec.Executor{
    Hosts:    []string{"host1.example.com", "host2.example.com"},
    Cmd:      "echo 'Hello from remote host'",
    Username: "username",
}

// Execute the command on all hosts
results := executor.Execute()

// Process the results
for _, result := range results {
    if result.Err != nil {
        fmt.Printf("Error on %s: %v\n", result.Host, result.Err)
    } else {
        fmt.Printf("Output from %s: %s\n", result.Host, result.Stdout)
    }
}
```

## Requirements

- SSH agent running locally with loaded keys
- SSH_AUTH_SOCK environment variable set
- Network access to remote hosts

## Testing

Run tests with a specific remote host:

```bash
TEST_EXEC_HOSTS=host.example.com TEST_EXEC_USERNAME=username go test -v ./pkg/exec
```