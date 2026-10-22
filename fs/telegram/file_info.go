package telegram

import (
	"os"
	"path/filepath"
	"time"
)

// FileData is a simple structure to store file information and implement os.FileInfo interface
type FileData struct {
	name    string
	dir     bool
	mode    os.FileMode
	modtime time.Time
	size    int64
	content []byte
}

// FileInfo implements the os.FileInfo interface on top of FileData
type FileInfo struct {
	*FileData
}

// Name returns the base name of the file
func (s *FileInfo) Name() string {
	_, name := filepath.Split(s.name)
	return name
}

// Mode returns the file mode bits
func (s *FileInfo) Mode() os.FileMode { return s.mode }

// ModTime returns the modification time
func (s *FileInfo) ModTime() time.Time { return s.modtime }

// IsDir tells if the file is a directory
func (s *FileInfo) IsDir() bool { return s.dir }

// Sys returns the underlying data source (always nil)
func (s *FileInfo) Sys() any { return nil }

// Size returns the size of the file
func (s *FileInfo) Size() int64 {
	if s.IsDir() {
		return int64(42)
	}
	return s.size
}
