package ssh

import (
	"net"
	"os"
	"strconv"
	"time"

	"vpn-manager/internal/data"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func dial(auth []ssh.AuthMethod) (*ssh.Client, error) {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return nil, err
	}

	sshUsername := settings.SSHUsername

	ConnectionTimeout := settings.ConnectionTimeout
	num, err := strconv.Atoi(ConnectionTimeout)
	if err != nil || num <= 0 {
		return nil, err
	}

	serverAddress := settings.ServerAddress
	sshPort := settings.SSHPort

	// All Server Addresses
	fullServer := net.JoinHostPort(serverAddress, sshPort)

	// Verify that the server is genuine
	hostKeyCallback, err := knownhosts.New(settings.SSHKnownHostsPath)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            sshUsername,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(num) * time.Second,
	}

	// Connect to the specified server
	client, err := ssh.Dial("tcp", fullServer, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ConnectServer() (*ssh.Client, error) {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return nil, err
	}

	// Open a file based on the SSH key path setting
	key, err := os.ReadFile(settings.SSHKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return dial([]ssh.AuthMethod{ssh.PublicKeys(signer)})
}
func ConnectWithPassword(password string) (*ssh.Client, error) {
	return dial([]ssh.AuthMethod{
		ssh.Password(password),
		// Some servers only allow keyboard-interactive instead of password
		ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = password
			}
			return answers, nil
		}),
	})
}

func RunCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
