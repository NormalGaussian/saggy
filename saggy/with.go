package saggy

import (
	"os"
	"os/exec"
	"strings"
)

func With(target string, command []string, mode string) error {
	if is_dir, s_err := isDir(target); s_err != nil {
		return s_err
	} else if is_dir {
		return withFolder(target, command, mode)
	} else {
		return withFile(target, command, mode)
	}
}

func withFile(file string, command []string, mode string) error {
	tmpFile, s_err := createTempFile()
	if s_err != nil {
		return NewSaggyError("Failed to create temporary file", s_err)
	}
	defer os.Remove(tmpFile)

	if s_err := DecryptFile(file, tmpFile); s_err != nil {
		return NewSaggyError("Failed to decrypt file", s_err)
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
		return NewExecutionError("Failed to run command", "", cmd.ProcessState.ExitCode(), cmd.Path, cmd.Args, cmd.Dir)
	}

	if mode == "write" {
		return EncryptFile(tmpFile, file)
	}
	return nil
}

func withFolder(folder string, command []string, mode string) error {
	tmpFolder, s_err := createTempDir()
	if s_err != nil {
		return NewSaggyError("Failed to create temporary directory", s_err)
	}
	defer os.RemoveAll(tmpFolder)

	if s_err := DecryptFolder(folder, tmpFolder); s_err != nil {
		return NewSaggyError("Failed to decrypt folder", s_err)
	}

	// Substitute {} with the temporary folder and quotes all args
	for i := range command {
		command[i] = strings.ReplaceAll(command[i], "{}", tmpFolder)
	}
	subcommand := strings.Join(command, " ")

	cmd := exec.Command("sh", "-c", subcommand)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return NewExecutionError("Failed to run command", "", cmd.ProcessState.ExitCode(), cmd.Path, cmd.Args, cmd.Dir)
	}

	if mode == "write" {
		return EncryptFolder(tmpFolder, folder)
	}
	return nil
}
