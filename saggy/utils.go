package saggy

import (
	"os/exec"
	"strings"
	"fmt"
	"os"
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
	// ...existing code...
	return ""
}

func getSopsifiedDirname(dir string) string {
	// ...existing code...
	return ""
}

func createTempFile() string {
	tmpFile, err := os.CreateTemp("", "saggy")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create temporary file:", err)
		os.Exit(1)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

func createTempDir() string {
	tmpDir, err := os.MkdirTemp("", "saggy")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create temporary directory:", err)
		os.Exit(1)
	}
	return tmpDir
}

func runCommand(command, target string) {
	cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "{}", target))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to run command:", err)
		os.Exit(1)
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to stat path:", err)
		os.Exit(1)
	}
	return info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to stat path:", err)
		os.Exit(1)
	}
	return !info.IsDir()
}