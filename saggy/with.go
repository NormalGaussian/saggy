package saggy

import (
	"os"
	"os/exec"
	"strings"
)

func With(keys *Keys, target string, command []string, mode string) error {
	if mode == "write" && keys.publicKeys == nil {
		return NewSaggyError("Cannot write - no public keys provided", nil)
	}

	if is_dir, err := isDir(target); err != nil {
		return err
	} else if is_dir {
		return withFolder(keys, target, command, mode)
	} else {
		return withFile(keys, target, command, mode)
	}
}

func withFile(keys *Keys, file string, command []string, mode string) error {
	tmpFile, s_err := createTempFile()
	if s_err != nil {
		return NewSaggyError("Failed to create temporary file", s_err)
	}
	defer os.Remove(tmpFile)

	if err := DecryptFile(keys.DecryptKey, file, tmpFile); err != nil {
		return err
	}
	if mode == "write" {
		defer EncryptFile(keys.EncryptKeys, tmpFile, file)
	}

	// Substitute {} with the temporary file
	for i := range command {
		command[i] = strings.ReplaceAll(command[i], "{}", tmpFile)
	}
	subcommand := strings.Join(command, " ")

	cmd := exec.Command("sh", "-c", subcommand)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		exactError := NewCommandError("Failed to run command", "", cmd)
		return NewSilentError(exactError, cmd.ProcessState.ExitCode())
	}

	return nil
}

func withFolder(keys *Keys, folder string, command []string, mode string) error {
	tmpFolder, err := createTempDir()
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpFolder)

	if err := DecryptFolder(keys.DecryptKey, folder, tmpFolder); err != nil {
		return err
	}
	if mode == "write" {
		defer EncryptFolder(keys.EncryptKeys, tmpFolder, folder)
	}

	// Substitute {} with the temporary folder
	for i := range command {
		command[i] = strings.ReplaceAll(command[i], "{}", tmpFolder)
	}
	subcommand := strings.Join(command, " ")

	// Run the command
	cmd := exec.Command("sh", "-c", subcommand)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		exactError := NewCommandError("Failed to run command", "", cmd)
		return NewSilentError(exactError, cmd.ProcessState.ExitCode())
	}

	return nil
}
