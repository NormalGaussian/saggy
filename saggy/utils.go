package saggy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func unsopsifyFilename(file string) string {
	if len(file) > 5 && file[len(file)-5:] == ".sops" {
		return file[:len(file)-5]
	}
	return file
}

func endWithSlash(path string) string {
	if path[len(path)-1] != '/' {
		return path + "/"
	}
	return path
}

func isSopsifiedFilename(file string) bool {
	return len(file) > 5 && file[len(file)-5:] == ".sops"
}

func getSopsifiedFilename(file string) string {
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	ext := filepath.Ext(base)

	return filepath.Join(dir, base[:len(base) - len(ext)]+".sops")
}

func getSopsifiedDirname(dir string) string {
	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	return dir + ".sops"
}

func createTempFile() (string, error) {
	tmpFile, err := os.CreateTemp("", "saggy")
	if err != nil {
		return "", NewSaggyError_skipFrames("Failed to create temporary file", err, nil, 2)
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func createTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "saggy")
	if err != nil {
		return "", NewSaggyError_skipFrames("Failed to create temporary directory", err, nil, 2)
	}
	return tmpDir, nil
}

func runCommand(command, target string) error {
	cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "{}", target))
	if output, err := cmd.Output(); err != nil {
		return NewExecutionError("Failed to run command", string(output), cmd.ProcessState.ExitCode(), cmd.Path, cmd.Args, cmd.Dir)
	}
	return nil
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, NewSaggyErrorWithMeta("Failed to stat path", err, info)
	}
	return info.IsDir(), nil
}

func isFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, NewSaggyErrorWithMeta("Failed to stat path", err, info)
	}
	return !info.IsDir(), nil
}
