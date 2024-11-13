package saggy

import (
	"os"
	"os/exec"
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
	if strings.Contains(file, ".") {
		parts := strings.Split(file, ".")
		return strings.Join(parts[:len(parts)-1], ".") + ".sops." + parts[len(parts)-1]
	}
	return file + ".sops"
}

func getSopsifiedDirname(dir string) string {
	return dir + ".sops"
}

func createTempFile() (string, *SaggyError) {
	tmpFile, err := os.CreateTemp("", "saggy")
	if err != nil {
		return "", &SaggyError{"Failed to create temporary file", err}
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func createTempDir() (string, *SaggyError) {
	tmpDir, err := os.MkdirTemp("", "saggy")
	if err != nil {
		return "", &SaggyError{"Failed to create temporary directory", err}
	}
	return tmpDir, nil
}

func runCommand(command, target string) *SaggyError {
	cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "{}", target))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &SaggyError{"Failed to run command", err}
	}
	return nil
}

func isDir(path string) (bool, *SaggyError) {
	info, err := os.Stat(path)
	if err != nil {
		return false, &SaggyError{"Failed to stat path", err}
	}
	return info.IsDir(), nil
}

func isFile(path string) (bool, *SaggyError) {
	info, err := os.Stat(path)
	if err != nil {
		return false, &SaggyError{"Failed to stat path", err}
	}
	return !info.IsDir(), nil
}
