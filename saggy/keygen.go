package saggy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"filippo.io/age"
)

func Keygen() error {
	if stat, err := os.Stat(keyFile); err == nil && stat != nil {
		return NewSaggyError("Key already exists - to generate a new key, delete the existing key\n"+
			"1. Decrypt the folders\n"+
			"  saggy decrypt <target> <destination>\n"+
			"2. Delete the key\n"+
			"  rm \"./secrets/age.key\"\n"+
			"2. Delete it from the public keys file\n"+
			"  vi \"./secrets/public-age-keys.json\"\n"+
			"3. Run this command again\n"+
			"  saggy keygen\n"+
			"4. Encrypt the folders\n"+
			"  saggy encrypt <target> <destination>\n", err)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewSaggyError("Unable to determine if the key file already exists", err)
	}

	if err := os.MkdirAll(filepath.Dir(keyFile), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	}

	var publicKey string
	if useBundledDependencies {
		cmd := exec.Command("age-keygen", "-o", keyFile)
		if output, err := cmd.CombinedOutput(); err != nil {
			return NewExecutionError("Failed to generate the key", string(output), cmd.ProcessState.ExitCode(), cmd.Path, cmd.Args, cmd.Dir)
		}

		publicKeyCmd := exec.Command("age-keygen", "-y", keyFile)
		publicKey_tmp, err := publicKeyCmd.Output()
		if err != nil {
			return NewSaggyError("Failed to get public key", err)
		}
		publicKey = string(publicKey_tmp)
	} else {
		// Generate the key
		k, err := age.GenerateX25519Identity()
		if err != nil {
			return NewSaggyError("Failed to generate the key", err)
		}

		// Extract the public key
		publicKey = k.Recipient().String()

		// Write the private key to the key file
		if f, err := os.OpenFile(keyFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
			return NewSaggyError("Failed to open key file for writing", err)
		} else {
			defer f.Close()
			fmt.Fprintf(f, "# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), publicKey, k)
		}
		
	}

	// Ensure the public keys file exists
	if _, err := os.Stat(publicKeysFile); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(publicKeysFile), 0755); err != nil {
			return NewSaggyError("Failed to create directory", err)
		}
		if err := os.WriteFile(publicKeysFile, []byte("{}"), 0644); err != nil {
			return NewSaggyError("Failed to create public keys file", err)
		}
	}

	// Add the public key to the public keys file
	publicKeys := make(map[string]string)
	file, err := os.ReadFile(publicKeysFile)
	if err != nil {
		return NewSaggyError("Failed to read public keys file", err)
	}
	if err := json.Unmarshal(file, &publicKeys); err != nil {
		return NewSaggyError("Failed to parse public keys file", err)
	}
	publicKeys[keyName] = publicKey
	updatedKeys, err := json.MarshalIndent(publicKeys, "", "  ")
	if err != nil {
		return NewSaggyError("Failed to serialize public keys", err)
	}
	if err := os.WriteFile(publicKeysFile, updatedKeys, 0644); err != nil {
		return NewSaggyError("Failed to write public keys file", err)
	}

	return nil
}
