package ssh

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func SendPublicKey(client *ssh.Client, publicKey []byte) error {
	// Create a new SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create a pipe to send the public key data to the remote server
	data, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to create stdin pipe: %w", err)
	}

	// Prepare the command to append the public key to the authorized_keys file
	remotePath := "$HOME/.ssh/authorized_keys"
	err = session.Start(fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> %s && chmod 600 %s", remotePath, remotePath))
	if err != nil {
		return fmt.Errorf("Failed to start remote command: %w", err)
	}

	// Write the public key data to the remote server
	if !bytes.HasSuffix(publicKey, []byte("\n")) {
		publicKey = append(publicKey, '\n')
	}

	_, err = data.Write(publicKey)
	if err != nil {
		return fmt.Errorf("Failed to write public key: %w", err)
	}

	// Close the stdin pipe to signal that we're done sending data
	err = data.Close()
	if err != nil {
		return fmt.Errorf("Failed to close stdin pipe: %w", err)
	}

	// Wait for the remote command to complete
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("Failed to complete remote command: %w", err)
	}

	return nil
}
