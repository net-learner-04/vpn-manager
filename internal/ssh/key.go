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
