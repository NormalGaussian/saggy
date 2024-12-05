package saggy

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// TODO: credit properly; taken and modified from go source
func safeCreateTempFile(nextName func() string, perms fs.FileMode) (*os.File, error) {
	try := 0
	for {
		f, err := os.OpenFile(nextName(), os.O_RDWR|os.O_CREATE|os.O_EXCL, perms)
		if errors.Is(err, fs.ErrExist) {
			if try++; try < 10000 {
				continue
			}
			return nil, err
		}
		return f, err
	}
}

func safeCreateAbsoluteTempFile(perms fs.FileMode) (*os.File, error) {
	return safeCreateTempFile(func() string {
		return filepath.Join(os.TempDir(), randomDecimalString())
	}, perms)
}

// TODO: credit properly; taken and modified from go source
func safeCreateRelativeTempFile(forFilename string, perms fs.FileMode) (*os.File, error) {
	dir := filepath.Dir(forFilename)
	basename := filepath.Base(forFilename)

	var prefix string
	if strings.HasPrefix(basename, ".") {
		prefix = basename
	} else {
		prefix = "." + basename
	}
	suffix := ".tmp"

	nextName := func() string {
		return filepath.Join(dir, prefix+randomDecimalString()+suffix)
	}

	return safeCreateTempFile(nextName, perms)
}

// TODO: credit properly; taken and modified from go source
func randomDecimalString() string {
	val := uint(rand.Uint32())
	if val == 0 { // avoid string allocation
		return "0"
	}
	var buf [20]byte // big enough for 64bit value base 10
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	// val < 10
	buf[i] = byte('0' + val)
	return string(buf[i:])
}

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

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, NewSaggyErrorWithMeta("Failed to stat path", err, struct {
			path string
		}{
			path: path,
		})
	}
	return info.IsDir(), nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get hostname:", err)
		os.Exit(1)
	}
	return hostname
}
