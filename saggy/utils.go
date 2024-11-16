package saggy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func unsopsifyFilename(file string) string {
	dirname := filepath.Dir(file)
	base := filepath.Base(file)
	parts := strings.Split(base, ".")
	if len(parts) == 2 && parts[1] == "sops" {
		return filepath.Join(dirname, parts[0])
	} else if len(parts) > 2 && parts[len(parts)-2] == "sops" {
		ext := filepath.Ext(base)
		return filepath.Join(dirname, base[:len(base)-len(ext)-5]+ext)
	}
	return file
}

func unsopsifyDirectory(dir string) string {
	filepath.Clean(dir)
	if len(dir) > 5 && dir[len(dir)-5:] == ".sops" {
		return dir[:len(dir)-5]
	}
	return dir
}

func isSopsifiedFilename(file string) bool {
	base := filepath.Base(file)
	parts := strings.Split(base, ".")
	if len(parts) == 2 && parts[1] == "sops" {
		return true
	} else if len(parts) > 2 && parts[len(parts)-2] == "sops" {
		return true
	}
	return false
}

func getSopsifiedFilename(file string) string {
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	ext := filepath.Ext(base)

	return filepath.Join(dir, base[:len(base)-len(ext)]+".sops"+ext)
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
