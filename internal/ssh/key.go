package ssh

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func GenerateSSHKey() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Generate a new Ed25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate SSH key pair: %w", err)
	}

	return privKey, pubKey, nil
}

func SavePrivateKey(privateKey ed25519.PrivateKey, filePath string) error {
	// Convert the private key to PKCS#8 format
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("Failed to marshal private key: %w", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}

	if err := os.WriteFile(filePath, pem.EncodeToMemory(block), 0600); err != nil {
		return fmt.Errorf("Failed to save private key: %w", err)
	}
	return nil
}

func SavePublicKey(publicKey ed25519.PublicKey, filePath string) error {
	// Convert the public key to OpenSSH format
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("Failed to convert SSH public key: %w", err)
	}

	// Add the public key file name as a comment
	comment := filepath.Base(filePath)
	keyData := bytes.TrimSpace(ssh.MarshalAuthorizedKey(sshPublicKey))
	keyData = append(keyData, []byte(" "+comment)...)
	keyData = append(keyData, '\n')

	// Save the public key to a file with appropriate permissions
	err = os.WriteFile(filePath, keyData, 0644)
	if err != nil {
		return fmt.Errorf("Failed to save SSH public key: %w", err)
	}
	return nil
}

// Register the SSH key on the server using a one-time password login.
// If a key already exists at privPath, it is reused instead of overwritten.
func RegisterKey(password, privPath, pubPath string) error {
	// Connect with the password (used only once, never stored)
	client, err := ConnectWithPassword(password)
	if err != nil {
		return err
	}
	defer client.Close()

	// Generate a new key pair only if there is no existing key
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		privKey, pubKey, err := GenerateSSHKey()
		if err != nil {
			return err
		}

		if err := SavePrivateKey(privKey, privPath); err != nil {
			return err
		}

		if err := SavePublicKey(pubKey, pubPath); err != nil {
			return err
		}
	}

	// Send the public key to the server's authorized_keys
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("Failed to read SSH public key: %w", err)
	}

	return SendPublicKey(client, pubBytes)
}

// Check whether the error is an authentication failure
func IsAuthError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unable to authenticate")
}
