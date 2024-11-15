package saggy

import (
	"os"
)

func With(target, command, mode string) error {
	if is_dir, s_err := isDir(target); s_err != nil {
		return s_err
	} else if is_dir {
		return withFolder(target, command, mode)
	} else {
		return withFile(target, command, mode)
	}
}

func withFile(file, command, mode string) error {
	tmpFile, s_err := createTempFile()
	if s_err != nil {
		return NewSaggyError("Failed to create temporary file", s_err)
	}
	defer os.Remove(tmpFile)

	if s_err := DecryptFile(file, tmpFile); s_err != nil {
		return NewSaggyError("Failed to decrypt file", s_err)
	}

	if s_err := runCommand(command, tmpFile); s_err != nil {
		return NewSaggyError("Failed to run command", s_err)
	}

	if mode == "write" {
		return EncryptFile(tmpFile, file)
	}
	return nil
}

func withFolder(folder, command, mode string) error {
	tmpFolder, s_err := createTempDir()
	if s_err != nil {
		return NewSaggyError("Failed to create temporary directory", s_err)
	}
	defer os.RemoveAll(tmpFolder)

	if s_err := DecryptFolder(folder, tmpFolder); s_err != nil {
		return NewSaggyError("Failed to decrypt folder", s_err)
	}

	if s_err := runCommand(command, tmpFolder); s_err != nil {
		return NewSaggyError("Failed to run command", s_err)
	}

	if mode == "write" {
		return EncryptFolder(tmpFolder, folder)
	}
	return nil
}
