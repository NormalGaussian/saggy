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

func Keygen_age_via_path() (key string, publicKey string, err error) {
	cmd := exec.Command("age-keygen")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", "", NewCommandError("Failed to generate the key", string(output), cmd)
	} else if cmd.ProcessState.ExitCode() != 0 {
		return "", "", NewCommandError("Failed to generate the key", string(output), cmd)
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
			return "", "", NewCommandError("Failed to find the public key in the output of age-keygen", keydata, cmd)
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
			return "", "", NewCommandError("Failed to find the key in the output of age-keygen", keydata, cmd)
		}

		return key, publicKey, nil
	}
}

func WriteKeyToFile(key string, publicKey string, keyFile string) error {
	if err := os.MkdirAll(filepath.Dir(keyFile), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	} else if f, err := os.OpenFile(keyFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		return NewSaggyError("Failed to open key file for writing", err)
	} else {
		defer f.Close()
		fmt.Fprintf(f, "# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), publicKey, key)
		return nil
	}
}

func WritePublicKeyToFile(publicKeysFile string, keyName string, publicKey string) error {
	if err := os.MkdirAll(filepath.Dir(publicKeysFile), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	} else if f, err := os.OpenFile(publicKeysFile, os.O_CREATE|os.O_RDWR, 0644); err != nil {
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
		keys[keyName] = publicKey

		// Write the keys back to the file
		tempFile, err := os.CreateTemp(filepath.Dir(publicKeysFile), "public-keys-*.json")
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

		if err := os.Rename(tempFile.Name(), publicKeysFile); err != nil {
			return NewSaggyError("Failed to rename temporary file to public keys file", err)
		}

		return nil
	}
}

func Keygen_age_via_import() (key string, publicKey string, err error) {
	// Generate the key
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", NewSaggyError("Failed to generate the key", err)
	}

	// Extract the public key
	publicKey = k.Recipient().String()

	return k.String(), publicKey, nil
}

func Keygen() (key string, publicKey string, err error) {
	if useBundledDependencies {
		return Keygen_age_via_import()
	} else {
		return Keygen_age_via_path()
	}
}

func KeygenToStdout(format string) error {
	key, publicKey, err := Keygen()
	if err != nil {
		return err
	}

	switch format {
	case "age":
		if _, err := fmt.Printf("# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), publicKey, key); err != nil {
			return NewSaggyError("Failed to write key to stdout", err)
		}
	case "json":
		if _, err := fmt.Printf("{\n  \"key\": \"%s\",\n  \"publicKey\": \"%s\"\n}\n", key, publicKey); err != nil {
			return NewSaggyError("Failed to write key to stdout", err)
		}
	default:
		return NewSaggyError("Invalid format", nil)
	}

	return nil
}

func KeygenToFile(keyFile string, publicKeysFile string) error {
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

	key, publicKey, err := Keygen()
	if err != nil {
		return err
	}

	if err := WriteKeyToFile(key, publicKey, keyFile); err != nil {
		return err
	}

	if err := WritePublicKeyToFile(publicKeysFile, keyName, publicKey); err != nil {
		return err
	}

	return nil
}
