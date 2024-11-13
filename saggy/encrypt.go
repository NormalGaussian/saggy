package saggy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Encrypt(from, to string) {
	if isFile(from) {
		EncryptFile(from, to)
	} else {
		EncryptFolder(from, to)
	}
}

func EncryptFile(from, to string) {
	if to == "" {
		to = getSopsifiedFilename(from)
	}

	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create directory:", err)
		os.Exit(1)
	}

	cmd := exec.Command("sops", "--encrypt", agePublicKeys, from)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to encrypt file:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(to, output, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write encrypted file:", err)
		os.Exit(1)
	}
}

func EncryptFolder(from, to string) {
	from = endWithSlash(from)
	to = endWithSlash(to)

	if to == "" {
		to = getSopsifiedDirname(from)
	}

	err := filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(from, path)
			if err != nil {
				return err
			}
			encryptedFile := getSopsifiedFilename(relPath)
			if err := os.MkdirAll(filepath.Dir(to+encryptedFile), 0755); err != nil {
				return err
			}
			cmd := exec.Command("sops", "--encrypt", agePublicKeys, path)
			output, err := cmd.Output()
			if err != nil {
				return err
			}
			if err := os.WriteFile(to+encryptedFile, output, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to encrypt folder:", err)
		os.Exit(1)
	}
}
