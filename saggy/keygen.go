package saggy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
)

func Keygen_age_via_path(keys *GenerateKeys) error {
	cmd := exec.Command("age-keygen")
	if output, err := cmd.CombinedOutput(); err != nil {
		return NewCommandError("Failed to generate the key", string(output), cmd)
	} else if cmd.ProcessState.ExitCode() != 0 {
		return NewCommandError("Failed to generate the key", string(output), cmd)
	} else {
		keydata := string(output)
		lines := strings.Split(keydata, "\n")

		// Extract the public key
		publicKey := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "# public key: ") {
				publicKey = strings.TrimPrefix(line, "# public key: ")
				break
			}
		}
		if publicKey == "" {
			return NewCommandError("Failed to find the public key in the output of age-keygen", keydata, cmd)
		}

		// Extract the key
		key := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
				key = line
				break
			}
		}
		if key == "" {
			return NewCommandError("Failed to find the key in the output of age-keygen", keydata, cmd)
		}

		keys.privateKey = key
		keys.publicKey = publicKey

		return nil
	}
}

func WriteKeyToFileNames(keys *GenerateKeys, keyFileNames *KeyFileNames) error {
	if err := os.MkdirAll(filepath.Dir(keyFileNames.privateKeyFilepath), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	} else if f, err := os.OpenFile(keyFileNames.privateKeyFilepath, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		return NewSaggyError("Failed to open key file for writing", err)
	} else {
		defer f.Close()
		fmt.Fprintf(f, "# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), keys.publicKey, keys.privateKey)
		return nil
	}
}

func WritePublicKeyToFileNames(generatedKeys *GenerateKeys, keyFileNames *KeyFileNames, keyName string) error {
	if err := os.MkdirAll(filepath.Dir(keyFileNames.publicKeysFilepath), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	} else if f, err := os.OpenFile(keyFileNames.publicKeysFilepath, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return NewSaggyError("Failed to open public keys file", err)
	} else {
		defer f.Close()

		// Read existing keys
		keys := make(map[string]string)
		if fileInfo, err := f.Stat(); err != nil {
			return NewSaggyError("Failed to get file info", err)
		} else if fileInfo.Size() > 0 {
			if err := json.NewDecoder(f).Decode(&keys); err != nil {
				return NewSaggyError("Failed to parse public keys file", err)
			}
		}

		// Add the new key
		keys[keyName] = generatedKeys.publicKey

		// Write the keys back to the file
		tempFile, err := os.CreateTemp(filepath.Dir(keyFileNames.publicKeysFilepath), "public-keys-*.json")
		if err != nil {
			return NewSaggyError("Failed to create temporary file", err)
		}
		defer os.Remove(tempFile.Name())

		if err := json.NewEncoder(tempFile).Encode(keys); err != nil {
			return NewSaggyError("Failed to encode public keys to temporary file", err)
		}

		if err := tempFile.Close(); err != nil {
			return NewSaggyError("Failed to close temporary file", err)
		}

		if err := os.Rename(tempFile.Name(), keyFileNames.publicKeysFilepath); err != nil {
			return NewSaggyError("Failed to rename temporary file to public keys file", err)
		}

		return nil
	}
}

func Keygen_age_via_import(keys *GenerateKeys) (err error) {
	// Generate the key
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return NewSaggyError("Failed to generate the key", err)
	}

	keys.publicKey = k.Recipient().String()
	keys.privateKey = k.String()

	return nil
}

func Keygen(keys *GenerateKeys) (err error) {
	if useBundledDependencies {
		return Keygen_age_via_import(keys)
	} else {
		return Keygen_age_via_path(keys)
	}
}

func KeygenToStdout(format string) error {
	// Generate the keys
	keys := &GenerateKeys{}
	err := Keygen(keys)
	if err != nil {
		return err
	}

	switch format {
	case "age":
		if _, err := fmt.Printf("# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), keys.publicKey, keys.privateKey); err != nil {
			return NewSaggyError("Failed to write key to stdout", err)
		}
	case "json":
		if _, err := fmt.Printf("{\n  \"key\": \"%s\",\n  \"publicKey\": \"%s\"\n}\n", keys.privateKey, keys.publicKey); err != nil {
			return NewSaggyError("Failed to write key to stdout", err)
		}
	default:
		return NewSaggyError("Invalid format", nil)
	}

	return nil
}

func KeygenToFile(keyFileNames *KeyFileNames, keyName string) error {
	// Guards
	if keyFileNames.privateKeyFilepath == "" {
		return NewSaggyError("Private key file path is not set", nil)
	} else if keyFileNames.publicKeysFilepath == "" {
		return NewSaggyError("Public key file path is not set", nil)
	} else if keyName == "" {
		return NewSaggyError("Key name is not set", nil)
	}

	// Verify the keyfile does not already exist
	if stat, err := os.Stat(keyFileNames.privateKeyFilepath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewSaggyError("Unable to determine if the key file already exists", err)
	} else if stat != nil {
		// file (or dir) exists
		return NewSaggyError("Key already exists - to generate a new key, delete the existing key\n"+ROTATE_KEY_GUIDE, err)
	}

	// Generate keys
	keys := &GenerateKeys{}
	if err := Keygen(keys); err != nil {
		return err
	}

	// Write the keys to the files
	if err := WriteKeyToFileNames(keys, keyFileNames); err != nil {
		return err
	}
	if err := WritePublicKeyToFileNames(keys, keyFileNames, keyName); err != nil {
		return err
	}

	return nil
}
