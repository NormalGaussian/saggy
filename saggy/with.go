package saggy

import (
	"fmt"
	"os"
)

func With(target, command, mode string) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "File or folder does not exist:", target)
		os.Exit(1)
	}

	if isDir(target) {
		withFolder(target, command, mode)
	} else {
		withFile(target, command, mode)
	}
}

func withFile(file, command, mode string) {
	tmpFile := createTempFile()
	defer os.Remove(tmpFile)

	DecryptFile(file, tmpFile)
	runCommand(command, tmpFile)

	if mode == "write" {
		EncryptFile(tmpFile, file)
	}
}

func withFolder(folder, command, mode string) {
	tmpFolder := createTempDir()
	defer os.RemoveAll(tmpFolder)

	DecryptFolder(folder, tmpFolder)
	runCommand(command, tmpFolder)

	if mode == "write" {
		EncryptFolder(tmpFolder, folder)
	}
}
