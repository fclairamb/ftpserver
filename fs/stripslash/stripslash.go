package stripslash

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"time"
)

type StripSlashPathFs struct {
	source afero.Fs
	start  int
}

type StripSlashPathFile struct {
	afero.File
	start int
}

func (f *StripSlashPathFile) Name() string {
	return f.File.Name()[f.start:]
}

func NewStripSlashPathFs(source afero.Fs, start int) afero.Fs {
	return &StripSlashPathFs{source: source, start: start}
}

// on a afero.File outside the base path it returns the given afero.File name and an error,
// else the given afero.File with the base path prepended
func (b *StripSlashPathFs) RealPath(name string) (path string, err error) {
	if len(name) > b.start {
		return "", fmt.Errorf("path needs to at least %d chars long", b.start)
	}

	return name[b.start:], nil
}

func (b *StripSlashPathFs) Chtimes(name string, atime, mtime time.Time) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "chtimes", Path: name, Err: err}
	}
	return b.source.Chtimes(name, atime, mtime)
}

func (b *StripSlashPathFs) Chmod(name string, mode os.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "chmod", Path: name, Err: err}
	}
	return b.source.Chmod(name, mode)
}

func (b *StripSlashPathFs) Name() string {
	return "StripSlashPathFs"
}

func (b *StripSlashPathFs) Stat(name string) (fi os.FileInfo, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}
	return b.source.Stat(name)
}

func (b *StripSlashPathFs) Rename(oldname, newname string) (err error) {
	if oldname, err = b.RealPath(oldname); err != nil {
		return &os.PathError{Op: "rename", Path: oldname, Err: err}
	}
	if newname, err = b.RealPath(newname); err != nil {
		return &os.PathError{Op: "rename", Path: newname, Err: err}
	}
	return b.source.Rename(oldname, newname)
}

func (b *StripSlashPathFs) RemoveAll(name string) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "remove_all", Path: name, Err: err}
	}
	return b.source.RemoveAll(name)
}

func (b *StripSlashPathFs) Remove(name string) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "remove", Path: name, Err: err}
	}
	return b.source.Remove(name)
}

func (b *StripSlashPathFs) OpenFile(name string, flag int, mode os.FileMode) (f afero.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "openfile", Path: name, Err: err}
	}
	sourcef, err := b.source.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}
	return &StripSlashPathFile{File: sourcef, start: b.start}, nil
}

func (b *StripSlashPathFs) Open(name string) (f afero.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	sourcef, err := b.source.Open(name)
	if err != nil {
		return nil, err
	}
	return &StripSlashPathFile{File: sourcef, start: b.start}, nil
}

func (b *StripSlashPathFs) Mkdir(name string, mode os.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return b.source.Mkdir(name, mode)
}

func (b *StripSlashPathFs) MkdirAll(name string, mode os.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return b.source.MkdirAll(name, mode)
}

func (b *StripSlashPathFs) Create(name string) (f afero.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: err}
	}
	sourcef, err := b.source.Create(name)
	if err != nil {
		return nil, err
	}
	return &StripSlashPathFile{File: sourcef, start: b.start}, nil
}

func (b *StripSlashPathFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	name, err := b.RealPath(name)
	if err != nil {
		return nil, false, &os.PathError{Op: "lstat", Path: name, Err: err}
	}
	if lstater, ok := b.source.(afero.Lstater); ok {
		return lstater.LstatIfPossible(name)
	}
	fi, err := b.source.Stat(name)
	return fi, false, err
}
