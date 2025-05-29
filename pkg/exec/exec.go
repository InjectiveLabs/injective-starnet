package exec

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Result struct {
	Host     string
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Executor struct for running commands on remote hosts
type Executor struct {
	Hosts    []string
	Cmd      string
	Username string // Username for SSH connections
}

// Execute runs the command on all hosts in parallel and returns the results
func (e *Executor) Execute() []Result {
	// Get local SSH agent, so we can use it to authenticate with the remote hosts
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return []Result{{Err: fmt.Errorf("failed to connect to SSH agent: %w", err)}}
	}
	defer sshAgent.Close()

	connClient := agent.NewClient(sshAgent)

	// Create SSH config with agent forwarding
	sshConfig := &ssh.ClientConfig{
		User:            e.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeysCallback(connClient.Signers)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	resultsCh := make(chan Result, len(e.Hosts))
	var wg sync.WaitGroup

	for _, host := range e.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			result := e.executeOnHost(host, sshConfig, connClient)
			resultsCh <- result
		}(host)
	}

	wg.Wait()
	close(resultsCh)

	var results []Result
	for result := range resultsCh {
		results = append(results, result)
	}

	return results
}

// executeOnHost runs the command on a single host
func (e *Executor) executeOnHost(host string, sshConfig *ssh.ClientConfig, connClient agent.Agent) Result {
	result := Result{
		Host: host,
	}

	// Connect to the remote host
	client, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		result.Err = fmt.Errorf("failed to dial: %w", err)
		return result
	}
	defer client.Close()

	// Create a new session
	session, err := client.NewSession()
	if err != nil {
		result.Err = fmt.Errorf("failed to create session: %w", err)
		return result
	}
	defer session.Close()

	// Set up agent forwarding
	if err := agent.ForwardToAgent(client, connClient); err != nil {
		result.Err = fmt.Errorf("failed to setup agent forwarding: %w", err)
		return result
	}
	if err := agent.RequestAgentForwarding(session); err != nil {
		result.Err = fmt.Errorf("failed to request agent forwarding: %w", err)
		return result
	}

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf strings.Builder
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Execute the command
	cmd := fmt.Sprintf("bash -l -c \"%s\"", e.Cmd)
	err = session.Run(cmd)
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			result.Err = fmt.Errorf("failed to run command: %w", err)
		}
	}

	return result
}
