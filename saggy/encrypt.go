package saggy

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

func Encrypt(keys *EncryptKeys, from, to string) error {
	if is_dir, err := isDir(from); err != nil {
		return err
	} else if is_dir {
		return EncryptFolder(keys, from, to)
	} else {
		return EncryptFile(keys, from, to)
	}
}

func EncryptFile(keys *EncryptKeys, from, to string) error {
	if to == "" {
		to = getSopsifiedFilename(from)
	}

	args := []string{"--encrypt"}
	for _, key := range *keys.publicKeys {
		args = append(args, "--age", key)
	}
	args = append(args, from, to)
	cmd := exec.Command("sops", args...)
	output, err := cmd.Output()
	if err != nil {
		return NewExecutionError("Failed to encrypt file", string(output), cmd.ProcessState.ExitCode(), cmd.Path, cmd.Args, cmd.Dir)
	}

	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return NewSaggyError("Failed to create directory", err)
	}

	if err := os.WriteFile(to, output, 0644); err != nil {
		return NewSaggyError("Failed to write encrypted file", err)
	}
	return nil
}

func EncryptFolder(keys *EncryptKeys, from, to string) error {
	from = filepath.Clean(from)
	if to == "" {
		to = getSopsifiedDirname(from)
	}

	err := filepath.WalkDir(from, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return NewSaggyError("Failed to walk directory", err)
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(from, path)
			if err != nil {
				return err
			}

			encryptedFile := filepath.Join(to, getSopsifiedFilename(relPath))
			if err := os.MkdirAll(filepath.Dir(encryptedFile), 0755); err != nil {
				return NewSaggyError("Failed to create directory", err)
			}

			args := []string{"--encrypt"}
			for _, key := range *keys.publicKeys {
				args = append(args, "--age", key)
			}
			args = append(args, path)
			cmd := exec.Command("sops", args...)
			output, err := cmd.Output()
			if err != nil {
				return NewSaggyError("Failed to encrypt file", err)
			}
			if err := os.WriteFile(encryptedFile, output, 0644); err != nil {
				return NewSaggyError("Failed to write encrypted file", err)
			}
		}
		return nil
	})
	if err != nil {
		saggyErr := &SaggyError{}
		if errors.As(err, &saggyErr) {
			return saggyErr
		}
		return NewSaggyError("Failed to walk directory", err)
	}
	return nil
}

func ReadAgePublicKeys(publicKeysFile string) ([]string, error) {
	if _, err := os.Stat(publicKeysFile); errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}

	data, err := os.ReadFile(publicKeysFile)
	if err != nil {
		return nil, NewSaggyError("Failed to read public keys file", err)
	}

	var keys map[string]string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, NewSaggyError("Failed to parse public keys file", err)
	}

	publicKeys := []string{}
	for _, value := range keys {
		publicKeys = append(publicKeys, value)
	}

	return publicKeys, nil
}
