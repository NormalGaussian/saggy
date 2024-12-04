package saggy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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

type KeyGenParameters struct {
	keyName string
	
	privateKeyWriter io.Writer
	privateKeyFile *os.File
	privateKeyFilepath string

	// Either age or json
	// Optional; if the public keys filename is provided
	privateKeyFormat string

	// Optional; if public keys file or filepath is provided this can be generated
	publicKeysWriter io.Writer
	// Optional; if the writer is provided and the reader is not provided the existing keys will not be preserved
	// Optional; if neither the writer nor the reader are provided the reader will be created from the publicKeysFile or publicKeysFilepath
	publicKeysReader io.Reader

	// Optional; if both the Writer and the format are provided
	publicKeysFile *os.File

	// Optional; if the public keys filename is provided
	publicKeysFilepath string

	// Currently only supports json
	// Optional; if the public keys filepath is provided it will be inferred
	publicKeysFormat string
}

type KeyGenParametersIO struct {
	keyName string
	privateKeyFormat string
	writePrivateKey func (data []byte) error
	publicKeysFormat string
	readPublicKeys func () ([]byte, error)
	writePublicKeys func ([]byte) error
}

func safeReplaceFile(file os.File, data []byte) error {
	tmpname := file.Name() + ".*.tmp"
	tmpfile, err := os.Create(tmpname)
	if err != nil {
		return NewSaggyError("Failed to create temporary file", err)
	}
	defer os.Remove(tmpname)

	if _, err := tmpfile.Write(data); err != nil {
		return NewSaggyError("Failed to write to temporary file", err)
	}
	if err := tmpfile.Close(); err != nil {
		return NewSaggyError("Failed to close temporary file", err)
	}

	if err := os.Rename(tmpname, file.Name()); err != nil {
		return NewSaggyError("Failed to rename temporary file", err)
	}

	return nil
}

func safeOpenFile(filename string, flags int, perm fs.FileMode) (*os.File, error) {
	dirname := filepath.Dir(filename)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return nil, NewSaggyError("Failed to create parent directories", err)
	}
	return os.OpenFile(filename, flags | os.O_CREATE, perm)
}

func safeReplaceFilename(filename string, data []byte) error {
	directory := filepath.Dir(filename)
	basename := filepath.Base(filename)
	tmpname := filepath.Join(directory, "." + basename + ".*.tmp")

	// Ensure the directory exists
	if err := os.MkdirAll(directory, 0755); err != nil {
		return NewSaggyError("Failed to create parent directories directory", err)
	}

	// Create the temporary file
	tempFile, err := os.CreateTemp(directory, tmpname)
	if err != nil {
		return NewSaggyError("Failed to create temporary file", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the data to the temporary file
	if _, err := tempFile.Write(data); err != nil {
		return NewSaggyError("Failed to write to temporary file", err)
	}
	if err := tempFile.Close(); err != nil {
		return NewSaggyError("Failed to close temporary file", err)
	}

	// Rename the temporary file to the target file
	if err := os.Rename(tempFile.Name(), filename); err != nil {
		return NewSaggyError("Failed to rename temporary file", err)
	}

	return nil
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

	var writePrivateKey func (data []byte) error

	if parameters.privateKeyWriter != nil {
		writePrivateKey = func(data []byte) error {
			_, err := parameters.privateKeyWriter.Write(data)
			return err
		}
	} else {

		privateKeyFile := parameters.privateKeyFile
		if privateKeyFile == nil && parameters.privateKeyFilepath != "" {
			if f, err := safeOpenFile(parameters.privateKeyFilepath, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				return NewSaggyError("Failed to open private key file", err)
			} else {
				defer f.Close()
				privateKeyFile = f
			}
		}

		if privateKeyFile != nil {
			writePrivateKey = func(data []byte) error {
				return safeReplaceFile(*privateKeyFile, data)
			}
		}
	}

	var writePublicKeys func (data []byte) error
	var readPublicKeys func () ([]byte, error)

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
		publicKeysFile := parameters.publicKeysFile
		if publicKeysFile == nil && parameters.publicKeysFilepath != "" {
			if f, err := safeOpenFile(parameters.publicKeysFilepath, os.O_CREATE|os.O_RDWR, 0644); err != nil {
				return NewSaggyError("Failed to open public keys file", err)
			} else {
				defer f.Close()
				publicKeysFile = f
			}
		}

		if publicKeysFile != nil {
			writePublicKeys = func(data []byte) error {
				return safeReplaceFile(*publicKeysFile, data)
			}
			readPublicKeys = func() ([]byte, error) {
				if _, err := publicKeysFile.Seek(0, 0); err != nil {
					return nil, err
				}
				return io.ReadAll(publicKeysFile)
			}
		}
	}
	
	return KeygenToIO(&KeyGenParametersIO{
		keyName: keyName,
		privateKeyFormat: privateKeyFormat,
		writePrivateKey: writePrivateKey,
		publicKeysFormat: publicKeysFormat,
		readPublicKeys: readPublicKeys,
		writePublicKeys: writePublicKeys,
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
		keyName: keyName,
	})
}
