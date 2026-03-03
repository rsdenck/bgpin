package ssh

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client
type Client struct {
	host      string
	port      int
	username  string
	password  string
	sshClient *ssh.Client
	timeout   time.Duration
}

// Config holds SSH client configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient creates a new SSH client
func NewClient(config Config) (*Client, error) {
	if config.Port == 0 {
		config.Port = 22
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		host:     config.Host,
		port:     config.Port,
		username: config.Username,
		password: config.Password,
		timeout:  config.Timeout,
	}, nil
}

// Connect establishes SSH connection
func (c *Client) Connect(ctx context.Context) error {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.sshClient = client
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

// ExecuteCommand executes a command via SSH
func (c *Client) ExecuteCommand(ctx context.Context, command string) (string, error) {
	if c.sshClient == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	return string(output), nil
}

// ExecuteCommands executes multiple commands
func (c *Client) ExecuteCommands(ctx context.Context, commands []string) ([]string, error) {
	results := make([]string, 0, len(commands))
	
	for _, cmd := range commands {
		output, err := c.ExecuteCommand(ctx, cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
		results = append(results, output)
	}
	
	return results, nil
}
