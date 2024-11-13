package saggy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Decrypt(from, to string) {
	if isFile(from) {
		DecryptFile(from, to)
	} else {
		DecryptFolder(from, to)
	}
}

func DecryptFile(from, to string) {
	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create directory:", err)
		os.Exit(1)
	}

	cmd := exec.Command("sops", "--decrypt", from)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to decrypt file:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(to, output, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write decrypted file:", err)
		os.Exit(1)
	}
}

func DecryptFolder(from, to string) {
	from = endWithSlash(from)
	to = endWithSlash(to)

	err := filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
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
		fmt.Fprintln(os.Stderr, "Failed to decrypt folder:", err)
		os.Exit(1)
	}
}
