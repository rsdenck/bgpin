package netconf

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents a NETCONF client
type Client struct {
	host       string
	port       int
	username   string
	password   string
	sshClient  *ssh.Client
	timeout    time.Duration
}

// Config holds NETCONF client configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient creates a new NETCONF client
func NewClient(config Config) (*Client, error) {
	if config.Port == 0 {
		config.Port = 830 // Default NETCONF port
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

// Close closes the connection
func (c *Client) Close() error {
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

// ExecuteRPC executes a NETCONF RPC
func (c *Client) ExecuteRPC(ctx context.Context, rpc string) (string, error) {
	if c.sshClient == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Execute NETCONF subsystem
	if err := session.RequestSubsystem("netconf"); err != nil {
		return "", fmt.Errorf("failed to start netconf subsystem: %w", err)
	}

	// Send RPC and read response
	output, err := session.CombinedOutput(rpc)
	if err != nil {
		return "", fmt.Errorf("failed to execute RPC: %w", err)
	}

	return string(output), nil
}

// GetBGPNeighbors retrieves BGP neighbors (vendor-agnostic)
func (c *Client) GetBGPNeighbors(ctx context.Context) (string, error) {
	rpc := `<rpc><get-bgp-neighbor-information/></rpc>`
	return c.ExecuteRPC(ctx, rpc)
}

// GetBGPRoutes retrieves BGP routes
func (c *Client) GetBGPRoutes(ctx context.Context, prefix string) (string, error) {
	rpc := fmt.Sprintf(`<rpc><get-route-information><destination>%s</destination></get-route-information></rpc>`, prefix)
	return c.ExecuteRPC(ctx, rpc)
}
