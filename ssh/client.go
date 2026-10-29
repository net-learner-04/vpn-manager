package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"vpn-manager/internal/data"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func ConnectServer() (*ssh.Client, error) {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}

	sshUsername := settings.SSHUsername
	sshKeyPath := settings.SSHKeyPath

	ConnectionTimeout := settings.ConnectionTimeout
	num, err := strconv.Atoi(ConnectionTimeout)
	if err != nil || num <= 0 {
		return nil, fmt.Errorf("invalid connection timeout: %q", ConnectionTimeout)
	}

	serverAddress := settings.ServerAddress
	sshPort := settings.SSHPort

	// All Server Addresses
	fullServer := net.JoinHostPort(serverAddress, sshPort)

	// Open a file based on the SSH key path setting
	key, err := os.ReadFile(sshKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read SSH private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse SSH private key: %w", err)
	}

	// Verify that the server is genuine
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find user home directory: %w", err)
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            sshUsername,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(num) * time.Second,
	}

	// Connect to the specified server
	client, err := ssh.Dial("tcp", fullServer, config)
	if err != nil {
		return nil, fmt.Errorf("connect to SSH server: %w", err)
	}

	return client, nil
}

// Creating a Session for Executing Remote Commands
func RunCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("execute remote command: %w", err)
	}

	return string(output), nil
}
