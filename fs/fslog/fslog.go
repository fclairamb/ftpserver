// Package fslog provides an afero FS logging package
package fslog

import (
	"os"
	"time"

	"github.com/spf13/afero"

	log "github.com/fclairamb/go-log"
)

// File is a wrapper to log interactions around file accesses
type File struct {
	src           afero.File // Source file
	logger        log.Logger // Associated logger
	lengthRead    int        // Length read
	lengthWritten int        // Length written
}

// Fs is a wrapper to log interactions around file system accesses
type Fs struct {
	src    afero.Fs   // Source file system
	logger log.Logger // Associated logger
}

func logErr(logger log.Logger, err error) log.Logger {
	if err != nil {
		return logger.With("err", err, "failed", true)
	}

	return logger
}

// Create calls will logged
func (f *Fs) Create(name string) (afero.File, error) {
	src, err := f.src.Create(name)

	logErr(f.logger, err).Info("Created file")

	return &File{
		src:    src,
		logger: f.logger.With("fileName", name),
	}, err
}

// Mkdir calls will not be logged
func (f *Fs) Mkdir(name string, perm os.FileMode) error {
	return f.src.Mkdir(name, perm)
}

// MkdirAll calls will not be logged
func (f *Fs) MkdirAll(path string, perm os.FileMode) error {
	return f.src.MkdirAll(path, perm)
}

// Open calls will be logged
func (f *Fs) Open(name string) (afero.File, error) {
	src, err := f.src.Open(name)
	logger := f.logger.With("fileName", name)
	logErr(logger, err).Info("Opened file")

	return &File{
		src:    src,
		logger: logger,
	}, err
}

// OpenFile calls will be logged
func (f *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	src, err := f.src.OpenFile(name, flag, perm)

	logger := f.logger.With("fileName", name, "fileFlag", flag, "filePerm", perm)

	logErr(logger, err).Info("Opened file")

	return &File{
		src:    src,
		logger: logger,
	}, err
}

// Remove calls will be logged
func (f *Fs) Remove(name string) error {
	err := f.src.Remove(name)

	logErr(f.logger, err).Info("Deleted file", "fileName", name)

	return err
}

// RemoveAll calls will not be logged
func (f *Fs) RemoveAll(path string) error {
	return f.src.RemoveAll(path)
}

// Rename calls will not be logged
func (f *Fs) Rename(oldname, newname string) error {
	return f.src.Rename(oldname, newname)
}

// Stat calls will not be logged
func (f *Fs) Stat(name string) (os.FileInfo, error) {
	return f.src.Stat(name)
}

// Name calls will not be logged
func (f *Fs) Name() string {
	return f.src.Name()
}

// Chmod calls will not be logged
func (f *Fs) Chmod(name string, mode os.FileMode) error {
	return f.src.Chmod(name, mode)
}

// Chtimes calls will not be logged
func (f *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return f.src.Chtimes(name, atime, mtime)
}

// Chown calls will not be logged
func (f *Fs) Chown(name string, uid int, gid int) error {
	return f.src.Chown(name, uid, gid)
}

// Close calls will be logged
func (f *File) Close() error {
	err := f.src.Close()

	logger := logErr(f.logger, err)

	if f.lengthRead > 0 {
		logger = logger.With("lengthRead", f.lengthRead)
	}

	if f.lengthWritten > 0 {
		logger = logger.With("lengthWritten", f.lengthWritten)
	}

	logger.Info("Closed file")

	return err
}

// Read won't be logged
func (f *File) Read(p []byte) (int, error) {
	n, err := f.src.Read(p)

	if err == nil {
		f.lengthRead += n
	}

	return n, err
}

// ReadAt won't be logged
func (f *File) ReadAt(p []byte, off int64) (int, error) {
	n, err := f.src.ReadAt(p, off)

	if err == nil {
		f.lengthRead += n
	}

	return n, err
}

// Seek won't be logged
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.src.Seek(offset, whence)
}

// Write won't be logged
func (f *File) Write(p []byte) (int, error) {
	n, err := f.src.Write(p)

	if err == nil {
		f.lengthWritten += n
	}

	return n, err
}

// WriteAt won't be logged
func (f *File) WriteAt(p []byte, off int64) (int, error) {
	n, err := f.src.WriteAt(p, off)

	if err == nil {
		f.lengthWritten += n
	}

	return n, err
}

// Name won't be logged
func (f *File) Name() string {
	return f.src.Name()
}

// Readdir won't be logged
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	return f.src.Readdir(count)
}

// Readdirnames won't be logged
func (f *File) Readdirnames(n int) ([]string, error) {
	return f.src.Readdirnames(n)
}

// Stat won't be logged
func (f *File) Stat() (os.FileInfo, error) {
	return f.src.Stat()
}

// Sync won't be logged
func (f *File) Sync() error {
	return f.src.Sync()
}

// Truncate won't be logged
func (f *File) Truncate(size int64) error {
	return f.src.Truncate(size)
}

// WriteString won't be logged
func (f *File) WriteString(str string) (int, error) {
	n, err := f.src.WriteString(str)

	if err == nil {
		f.lengthWritten += n
	}

	return n, err
}

// LoadFS creates an instance with logging
func LoadFS(src afero.Fs, logger log.Logger) (afero.Fs, error) {
	return &Fs{
		src:    src,
		logger: logger,
	}, nil
}
