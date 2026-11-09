package ssh

import (
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"vpn-manager/internal/data"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func dial(serverAddress string, auth []ssh.AuthMethod) (*ssh.Client, error) {
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

	// An empty address falls back to the saved server address.
	if serverAddress == "" {
		serverAddress = settings.ServerAddress
	}
	sshPort := settings.SSHPort

	// All Server Addresses
	fullServer := net.JoinHostPort(serverAddress, sshPort)

	// Verify that the server is genuine. On first contact the host key is
	// returned as a HostKeyError so the UI can ask the user to approve it.
	baseCallback, err := knownhosts.New(settings.SSHKnownHostsPath)
	if err != nil {
		return nil, err
	}

	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := baseCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			// Unknown host key (first contact)
			return &HostKeyError{
				Address:     hostname,
				Fingerprint: ssh.FingerprintSHA256(key),
				Key:         key,
			}
		}

		return err
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
	return ConnectServerWithAddress("")
}

// ConnectServerWithAddress connects to the server using an explicit address.
// Pass an empty address to use the saved server address.
func ConnectServerWithAddress(serverAddress string) (*ssh.Client, error) {
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

	return dial(serverAddress, []ssh.AuthMethod{ssh.PublicKeys(signer)})
}
func ConnectWithPassword(password string) (*ssh.Client, error) {
	return dial("", []ssh.AuthMethod{
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

// HostKeyError reports an unknown server host key so the UI can ask the user
// to approve the first connection.
type HostKeyError struct {
	Address     string
	Fingerprint string
	Key         ssh.PublicKey
}

func (e *HostKeyError) Error() string {
	return "Unknown host key for " + e.Address + ": " + e.Fingerprint
}

// TrustHostKey appends the server key to the known_hosts file.
func TrustHostKey(address string, key ssh.PublicKey) error {
	settings, err := data.ReadSettingsFile()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(settings.SSHKnownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(knownhosts.Line([]string{address}, key))
	return err
}
