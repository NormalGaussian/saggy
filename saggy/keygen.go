package saggy

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

func Keygen() *SaggyError {
	if _, err := os.Stat(keyFile); err == nil {
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
			"  saggy encrypt <target> <destination>", err)
	}

	if err := os.MkdirAll(filepath.Dir(keyFile), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	}

	cmd := exec.Command("age-keygen", "-o", keyFile)
	if err := cmd.Run(); err != nil {
		return NewSaggyError("Failed to generate the key", err)
	}

	publicKeyCmd := exec.Command("age-keygen", "-y", keyFile)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		return NewSaggyError("Failed to get public key", err)
	}

	if _, err := os.Stat(publicKeysFile); os.IsNotExist(err) {
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
	publicKeys[filepath.Base(keyFile)] = string(publicKey)
	updatedKeys, err := json.MarshalIndent(publicKeys, "", "  ")
	if err != nil {
		return NewSaggyError("Failed to serialize public keys", err)
	}
	if err := os.WriteFile(publicKeysFile, updatedKeys, 0644); err != nil {
		return NewSaggyError("Failed to write public keys file", err)
	}

	return nil
}
