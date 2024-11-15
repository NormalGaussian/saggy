package saggy

import (
	"os"
	"os/exec"
	"path/filepath"
)

func Decrypt(from, to string) error {
	if is_dir, s_err := isDir(from); s_err != nil {
		return s_err
	} else if is_dir {
		return DecryptFolder(from, to)
	} else {
		return DecryptFile(from, to)
	}
}

func DecryptFile(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return NewSaggyError("Failed to create directory:", err)
	}

	cmd := exec.Command("sops", "--decrypt", from)
	cmd.Env = []string{"SOPS_AGE_KEY_FILE=" + keyFile}
	output, err := cmd.Output()

	if err != nil {
		return NewSaggyError("Failed to decrypt file:", err)
	}

	if err := os.WriteFile(to, output, 0644); err != nil {
		return NewSaggyError("Failed to write decrypted file:", err)
	}

	return nil
}

func DecryptFolder(from, to string) error {
	from = endWithSlash(from)
	to = endWithSlash(to)

	err := filepath.WalkDir(from, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			encryptedFile := path[len(from):]
			if !isSopsifiedFilename(encryptedFile) {
				return nil
			}
			decryptedFile := unsopsifyFilename(encryptedFile)
			if err := os.MkdirAll(filepath.Join(to, filepath.Dir(decryptedFile)), 0755); err != nil {
				return err
			}
			cmd := exec.Command("sops", "--decrypt", path)
			output, err := cmd.Output()
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(to, decryptedFile), output, 0644); err != nil {
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
