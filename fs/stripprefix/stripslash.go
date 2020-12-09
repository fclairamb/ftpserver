// Package stripprefix is a file-system abstraction layer to strip a part of the path
package stripprefix

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/afero"
)

// ErrBasePathTooShort is returned when the specified path is too short
var ErrBasePathTooShort = errors.New("path needs to at least as long as its prefix")

// Fs is a convenience afero.Fs to remove the prefix of a path
type Fs struct {
	source afero.Fs
	start  int
}

// File is the afero.File implementation
type File struct {
	afero.File
	start int
}

// NewStripPrefixFs is an internal FS implementation to remove a path prefix
// of a path when calling the underlying FS implementation.
func NewStripPrefixFs(source afero.Fs, start int) afero.Fs {
	return &Fs{source: source, start: start}
}

// Name of the file
func (f *File) Name() string {
	return f.File.Name()[f.start:]
}

// on a afero.File outside the base path it returns the given afero.File name and an error,
// else the given afero.File with the base path prepended
func (b *Fs) realPath(name string) (path string, err error) {
	if len(name) < b.start {
		return "", ErrBasePathTooShort
	}

	return name[b.start:], nil
}

// Chtimes changes the access and modification times of the named file
func (b *Fs) Chtimes(name string, atime, mtime time.Time) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "chtimes", Path: name, Err: err}
	}

	return b.source.Chtimes(name, atime, mtime)
}

// Chmod changes the mode of the named file to mode.
func (b *Fs) Chmod(name string, mode os.FileMode) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "chmod", Path: name, Err: err}
	}

	return b.source.Chmod(name, mode)
}

// Chown changes the mode of the named file to mode.
func (b *Fs) Chown(name string, uid int, gid int) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "chmod", Path: name, Err: err}
	}

	return b.source.Chown(name, uid, gid)
}

// Name of this FileSystem
func (b *Fs) Name() string {
	return "Fs"
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (b *Fs) Stat(name string) (fi os.FileInfo, err error) {
	if name, err = b.realPath(name); err != nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}

	return b.source.Stat(name)
}

// Rename renames a file.
func (b *Fs) Rename(oldname, newname string) (err error) {
	if oldname, err = b.realPath(oldname); err != nil {
		return &os.PathError{Op: "rename", Path: oldname, Err: err}
	}

	if newname, err = b.realPath(newname); err != nil {
		return &os.PathError{Op: "rename", Path: newname, Err: err}
	}

	return b.source.Rename(oldname, newname)
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (b *Fs) RemoveAll(name string) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "remove_all", Path: name, Err: err}
	}

	return b.source.RemoveAll(name)
}

// Remove removes a file identified by name, returning an error if any happens.
func (b *Fs) Remove(name string) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "remove", Path: name, Err: err}
	}

	return b.source.Remove(name)
}

// OpenFile opens a file using the given flags and the given mode.
func (b *Fs) OpenFile(name string, flag int, mode os.FileMode) (f afero.File, err error) {
	if name, err = b.realPath(name); err != nil {
		return nil, &os.PathError{Op: "openfile", Path: name, Err: err}
	}

	sourcef, err := b.source.OpenFile(name, flag, mode)

	if err != nil {
		return nil, err
	}

	return &File{File: sourcef, start: b.start}, nil
}

// Open opens a file, returning it or an error, if any happens.
func (b *Fs) Open(name string) (f afero.File, err error) {
	if name, err = b.realPath(name); err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}

	sourcef, err := b.source.Open(name)

	if err != nil {
		return nil, err
	}

	return &File{File: sourcef, start: b.start}, nil
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (b *Fs) Mkdir(name string, mode os.FileMode) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}

	return b.source.Mkdir(name, mode)
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (b *Fs) MkdirAll(name string, mode os.FileMode) (err error) {
	if name, err = b.realPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}

	return b.source.MkdirAll(name, mode)
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (b *Fs) Create(name string) (f afero.File, err error) {
	if name, err = b.realPath(name); err != nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: err}
	}

	sourcef, err := b.source.Create(name)

	if err != nil {
		return nil, err
	}

	return &File{File: sourcef, start: b.start}, nil
}

// LstatIfPossible implements afero.Lstater.LstatIfPossible
func (b *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	name, err := b.realPath(name)
	if err != nil {
		return nil, false, &os.PathError{Op: "lstat", Path: name, Err: err}
	}

	if lstater, ok := b.source.(afero.Lstater); ok {
		return lstater.LstatIfPossible(name)
	}

	fi, err := b.source.Stat(name)

	return fi, false, err
}
