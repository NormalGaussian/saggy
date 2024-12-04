package saggy

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type SafeWholeFileIO interface {
	// Write the data to the file, replacing the entire file
	Write(data []byte) error

	// Read the entire file
	Read() ([]byte, error)

	// Remove the file
	Remove() error
}

type SafeWholeFile struct {
	filename string
	perms fs.FileMode
	flags int
}

var defaultPermissions = fs.FileMode(0644)
var defaultFlags = os.O_RDWR

func NewSafeWholeFile(filename string, flags int, perms fs.FileMode) *SafeWholeFile {
	if perms == 0 {
		perms = defaultPermissions
	}

	if flags == 0 {
		flags = defaultFlags
	}

	return &SafeWholeFile{
		filename: filename,
		perms: perms,
		flags: flags,
	}
}

func (s *SafeWholeFile) Write(data []byte) error {
	// Check if the file is openable for writing
	if s.flags & os.O_WRONLY == 0 && s.flags & os.O_RDWR == 0 {
		return NewSaggyError("File is not openable for writing", nil)
	}
	
	// Check if there are file issues beyong not existing
	stat, err := os.Stat(s.filename)
	if err != nil && !os.IsNotExist(err) {
		return NewSaggyError("Failed to stat file", err)
	} else if err == nil && stat.IsDir() {
		return NewSaggyError("File already exists as a directory", nil)
	}

	// If the file exists, check we aren't trying to exclusively create it
	if err == nil && s.flags & os.O_CREATE != 0 && s.flags & os.O_EXCL != 0 {
		return NewSaggyError("File already exists", nil)
	}

	// Ensure parent directories exist
	dirname := filepath.Dir(s.filename)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return NewSaggyError("Failed to create parent directories", err)
	}

	// Create the temporary file
	tmpfile, err := safeCreateRelativeTempFile(s.filename, s.perms)
	if err != nil {
		return NewSaggyError("Failed to create temporary file", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write the data to the temporary file
	if _, err := tmpfile.Write(data); err != nil {
		return NewSaggyError("Failed to write to temporary file", err)
	}
	if err := tmpfile.Close(); err != nil {
		return NewSaggyError("Failed to close temporary file", err)
	}

	// Rename the temporary file to the target file
	if err := os.Rename(tmpfile.Name(), s.filename); err != nil {
		return NewSaggyError("Failed to rename temporary file", err)
	}

	return nil
}

func (s *SafeWholeFile) Read() ([]byte, error) {
	// Check if the file is openable for reading
	if s.flags & os.O_RDONLY == 0 && s.flags & os.O_RDWR == 0 {
		return nil, NewSaggyError("File is not openable for reading", nil)
	}

	// Check if the file exists
	stat, err := os.Stat(s.filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, NewSaggyError("Failed to stat file", err)
	} else if err == nil && stat.IsDir() {
		return nil, NewSaggyError("File already exists as a directory", nil)
	}

	// If the file doesn't exist, it has no data to read
	if errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil
	}

	// Read the data
	data, err := os.ReadFile(s.filename)
	if err != nil {
		return nil, NewSaggyError("Failed to read file", err)
	}

	return data, nil
}

func (s *SafeWholeFile) Remove() error {
	// Check if the file is openable for writing
	if s.flags & os.O_WRONLY == 0 {
		return NewSaggyError("File is not openable for writing", nil)
	}

	// Check if the file exists
	stat, err := os.Stat(s.filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewSaggyError("Failed to stat file", err)
	} else if err == nil && stat.IsDir() {
		return NewSaggyError("File already exists as a directory", nil)
	}

	// No file to remove
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	// Remove the file
	if err := os.Remove(s.filename); err != nil {
		return NewSaggyError("Failed to remove file", err)
	}

	return nil
}