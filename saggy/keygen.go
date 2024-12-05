package saggy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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

type KeyGenParameters struct {
	// How the key should be identified in the public keys file
	// Optional; required if the public keys filename or writer is provided, and the key format is not age
	keyName string

	//
	privateKeyWriter io.Writer

	// Where to write the private key
	// Optional; if the writer is provided the file will not be created
	privateKeyFilepath string

	// Either age or json
	// Optional; if the public keys filename is provided
	privateKeyFormat string

	// Optional; if public keys file or filepath is provided this can be generated
	publicKeysWriter io.Writer
	// Optional; if the writer is provided and the reader is not provided the existing keys will not be preserved
	// Optional; if neither the writer nor the reader are provided the reader will be created from the publicKeysFile or publicKeysFilepath
	publicKeysReader io.Reader

	// Optional; if the public keys filename is provided
	publicKeysFilepath string

	// Currently only supports json
	// Optional; if the public keys filepath is provided it will be inferred
	publicKeysFormat string
}

type KeyGenParametersIO struct {
	keyName          string
	privateKeyFormat string
	writePrivateKey  func(data []byte) error
	publicKeysFormat string
	readPublicKeys   func() ([]byte, error)
	writePublicKeys  func([]byte) error
}

func KeyGen_parameterised(parameters *KeyGenParameters) error {
	// The goal is arrange the parameters to get reader and writers for the private and public keys alongside the appropriate format

	// Determine the private key format
	privateKeyFormat := parameters.privateKeyFormat
	if privateKeyFormat == "" {
		// Determine the format from the file extension
		if parameters.privateKeyFilepath == "" {
			return NewSaggyError("Private key format is not set and there is no filepath from which to infer it", nil)
		}
		if strings.HasSuffix(parameters.privateKeyFilepath, ".age") {
			privateKeyFormat = "age"
		} else if strings.HasSuffix(parameters.privateKeyFilepath, ".json") {
			privateKeyFormat = "json"
		} else {
			return NewSaggyError("Private key format is not set and the filepath does not have a recognised extension from which to infer it", nil)
		}
	}
	if privateKeyFormat != "age" && privateKeyFormat != "json" {
		return NewSaggyError("Invalid private key format", nil)
	}

	// Determine the public keys format
	publicKeysFormat := parameters.publicKeysFormat
	if publicKeysFormat == "" {
		// Determine the format from the file extension
		if parameters.publicKeysFilepath != "" {
			if strings.HasSuffix(parameters.publicKeysFilepath, ".age") {
				publicKeysFormat = "age"
			} else if strings.HasSuffix(parameters.publicKeysFilepath, ".json") {
				publicKeysFormat = "json"
			} else {
				return NewSaggyError("Public keys format is not set and the filepath does not have a recognised extension from which to infer it", nil)
			}
		} else if privateKeyFormat == "age" {
			// If the private key format is age then we can default the public keys format to age, as the age format includes the public key
			publicKeysFormat = "age"
		} else {
			return NewSaggyError("Public keys format is not set and there is no filepath from which to infer it", nil)
		}
	}
	if publicKeysFormat != "age" && publicKeysFormat != "json" {
		return NewSaggyError("Invalid public keys format", nil)
	}

	// Determine the key name
	keyName := parameters.keyName
	if keyName == "" && ((parameters.privateKeyWriter != nil && privateKeyFormat == "json") || (parameters.publicKeysWriter != nil && publicKeysFormat == "json")) {
		return NewSaggyError("Key name is not set", nil)
	}

	// At this point everything needed is set

	var writePrivateKey func(data []byte) error

	if parameters.privateKeyWriter != nil {
		writePrivateKey = func(data []byte) error {
			_, err := parameters.privateKeyWriter.Write(data)
			return err
		}
	} else if parameters.privateKeyFilepath != "" {
		f := NewSafeWholeFile(parameters.privateKeyFilepath, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
		writePrivateKey = f.Write
	}

	var writePublicKeys func(data []byte) error
	var readPublicKeys func() ([]byte, error)

	if parameters.publicKeysWriter != nil {
		writePublicKeys = func(data []byte) error {
			_, err := parameters.publicKeysWriter.Write(data)
			return err
		}
	}
	if parameters.publicKeysReader != nil {
		readPublicKeys = func() ([]byte, error) {
			return io.ReadAll(parameters.publicKeysReader)
		}
	}
	if writePublicKeys == nil && readPublicKeys == nil {
		if parameters.publicKeysFilepath != "" {
			f := NewSafeWholeFile(parameters.publicKeysFilepath, os.O_CREATE|os.O_RDWR, 0644)
			writePublicKeys = f.Write
			readPublicKeys = f.Read
		}
	}

	return KeygenToIO(&KeyGenParametersIO{
		keyName:          keyName,
		privateKeyFormat: privateKeyFormat,
		writePrivateKey:  writePrivateKey,
		publicKeysFormat: publicKeysFormat,
		readPublicKeys:   readPublicKeys,
		writePublicKeys:  writePublicKeys,
	})
}

func KeygenToIO(opts *KeyGenParametersIO) error {
	// Generate the keys
	keys := &GenerateKeys{}
	err := Keygen(keys)
	if err != nil {
		return err
	}

	// Write the private keys to the io.Writer if provided
	if opts.writePrivateKey != nil {
		var data string
		switch opts.privateKeyFormat {
		case "age":
			data = fmt.Sprintf("# created: %s\n# public key: %s\n%s\n", time.Now().Format(time.RFC3339), keys.publicKey, keys.privateKey)
		case "json":
			data = fmt.Sprintf("{\n  \"key\": \"%s\",\n  \"publicKey\": \"%s\"\n}\n", keys.privateKey, keys.publicKey)
		default:
			return NewSaggyError("Invalid format", nil)
		}
		if err := opts.writePrivateKey([]byte(data)); err != nil {
			return err
		}
	}

	// Write the public keys to the io.Writer if provided, reading the existing keys from the io.Reader if provided
	if opts.writePublicKeys != nil {
		switch opts.publicKeysFormat {
		case "age":

			// No reader is required for age format

			data := fmt.Sprintf("%s\n", keys.publicKey)
			if err := opts.writePublicKeys([]byte(data)); err != nil {
				return err
			}

		case "json":

			publicKeys := make(map[string]string)

			// Read the existing keys
			if opts.readPublicKeys != nil {
				data, err := opts.readPublicKeys()
				if err != nil {
					return err
				}

				if len(data) > 0 {

					if err := json.Unmarshal(data, &publicKeys); err != nil {
						return NewSaggyError("Failed to parse public keys", err)
					}
				}
			}

			// Add the new key
			publicKeys[opts.keyName] = keys.publicKey

			// Write the keys
			if data, err := json.Marshal(publicKeys); err != nil {
				return NewSaggyError("Failed to marshal public keys", err)
			} else if err := opts.writePublicKeys(data); err != nil {
				return err
			}
		}
	}

	return nil
}

func KeygenToStdout(format string) error {
	return KeyGen_parameterised(&KeyGenParameters{
		privateKeyWriter: os.Stdout,
		privateKeyFormat: format,
	})
}

func KeygenToFile(privateKeyFilepath, publicKeysFilepath, keyName string) error {
	return KeyGen_parameterised(&KeyGenParameters{
		privateKeyFilepath: privateKeyFilepath,
		publicKeysFilepath: publicKeysFilepath,
		keyName:            keyName,
	})
}
