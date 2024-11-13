package saggy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Keygen() {
	if _, err := os.Stat(keyFile); err == nil {
		fmt.Fprintln(os.Stderr, "Key already exists - to generate a new key, delete the existing key")
		fmt.Fprintln(os.Stderr, "1. Decrypt the folders")
		fmt.Fprintln(os.Stderr, "  saggy decrypt <target> <destination>")
		fmt.Fprintln(os.Stderr, "2. Delete the key")
		fmt.Fprintln(os.Stderr, "  rm \"./secrets/age.key\"")
		fmt.Fprintln(os.Stderr, "2. Delete it from the public keys file")
		fmt.Fprintln(os.Stderr, "  vi \"./secrets/public-age-keys.json\"")
		fmt.Fprintln(os.Stderr, "3. Run this command again")
		fmt.Fprintln(os.Stderr, "  saggy keygen")
		fmt.Fprintln(os.Stderr, "4. Encrypt the folders")
		fmt.Fprintln(os.Stderr, "  saggy encrypt <target> <destination>")
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(keyFile), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create directory:", err)
		os.Exit(1)
	}

	cmd := exec.Command("age-keygen", "-o", keyFile)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to generate the key:", err)
		os.Exit(1)
	}

	publicKeyCmd := exec.Command("age-keygen", "-y", keyFile)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get public key:", err)
		os.Exit(1)
	}

	if _, err := os.Stat(publicKeysFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(publicKeysFile), 0755); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create directory:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(publicKeysFile, []byte("{}"), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create public keys file:", err)
			os.Exit(1)
		}
	}

	// Add the public key to the public keys file
	publicKeys := make(map[string]string)
	file, err := os.ReadFile(publicKeysFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read public keys file:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &publicKeys); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse public keys file:", err)
		os.Exit(1)
	}
	publicKeys[filepath.Base(keyFile)] = string(publicKey)
	updatedKeys, err := json.MarshalIndent(publicKeys, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to serialize public keys:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(publicKeysFile, updatedKeys, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write public keys file:", err)
		os.Exit(1)
	}
}