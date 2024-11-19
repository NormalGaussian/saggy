package saggy

import (
	"os"
	"os/exec"
	"path/filepath"
)

func Decrypt(keys *DecryptKey, from, to string) error {
	if is_dir, err := isDir(from); err != nil {
		return err
	} else if is_dir {
		return DecryptFolder(keys, from, to)
	} else {
		return DecryptFile(keys, from, to)
	}
}

func DecryptFile(keys *DecryptKey, from, to string) error {
	from = filepath.Clean(from)
	if to == "" {
		to = unsopsifyFilename(from)
	}

	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return NewSaggyError("Failed to create directory:", err)
	}

	cmd := exec.Command("sops", "--decrypt", from)
	cmd.Env = []string{"SOPS_AGE_KEY_FILE=" + keys.privateKeyFilepath}
	output, err := cmd.Output()

	if err != nil {
		return NewSaggyError("Failed to decrypt file:", err)
	}

	if err := os.WriteFile(to, output, 0644); err != nil {
		return NewSaggyError("Failed to write decrypted file:", err)
	}

	return nil
}

func DecryptFolder(keys *DecryptKey, from, to string) error {
	from = filepath.Clean(from)
	if to == "" {
		to = unsopsifyDirectory(from)
	}

	err := filepath.WalkDir(from, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			encryptedFile, err := filepath.Rel(from, path)
			if err != nil {
				return err
			}
			if !isSopsifiedFilename(encryptedFile) {
				return nil
			}
			decryptedFile := filepath.Join(to, unsopsifyFilename(encryptedFile))
			if err := os.MkdirAll(filepath.Dir(decryptedFile), 0755); err != nil {
				return err
			}
			cmd := exec.Command("sops", "--decrypt", path)
			cmd.Env = []string{"SOPS_AGE_KEY_FILE=" + keys.privateKeyFilepath}
			output, err := cmd.Output()
			if err != nil {
				return err
			}
			if err := os.WriteFile(decryptedFile, output, 0644); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return NewSaggyError("Failed to decrypt folder:", err)
	}
	return nil
}
