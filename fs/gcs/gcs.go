// Package gcs provides a Google Cloud Storage access layer
package gcs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/spf13/afero"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// gcsFS implements afero.Fs for Google Cloud Storage
type gcsFS struct {
	client *storage.Client
	bucket string
	ctx    context.Context
}

// LoadFs loads a GCS file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	bucket := access.Params["bucket"]
	if bucket == "" {
		return nil, errors.New("bucket parameter is required for GCS")
	}

	projectID := access.Params["project_id"]
	keyFile := access.Params["key_file"]

	ctx := context.Background()
	var client *storage.Client
	var err error

	// Create client with different authentication methods
	if keyFile != "" {
		// Use service account key file
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(keyFile))
	} else if projectID != "" {
		// Use default credentials with specific project ID
		client, err = storage.NewClient(ctx, option.WithScopes(storage.ScopeFullControl))
	} else {
		// Use default credentials
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &gcsFS{
		client: client,
		bucket: bucket,
		ctx:    ctx,
	}, nil
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (fs *gcsFS) Create(name string) (afero.File, error) {
	return &gcsFile{
		fs:   fs,
		name: name,
		mode: 0644,
	}, nil
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (fs *gcsFS) Mkdir(name string, perm os.FileMode) error {
	// GCS doesn't have real directories, this is a no-op
	return nil
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (fs *gcsFS) MkdirAll(path string, perm os.FileMode) error {
	// GCS doesn't have real directories, this is a no-op
	return nil
}

// Open opens a file, returning it or an error, if any happens.
func (fs *gcsFS) Open(name string) (afero.File, error) {
	obj := fs.client.Bucket(fs.bucket).Object(name)
	
	// Check if object exists
	_, err := obj.Attrs(fs.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, fmt.Errorf("file does not exist: %s", name)
		}
		return nil, err
	}

	return &gcsFile{
		fs:   fs,
		name: name,
		mode: 0644,
	}, nil
}

// OpenFile opens a file using the given flags and the given mode.
func (fs *gcsFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return &gcsFile{
		fs:   fs,
		name: name,
		mode: perm,
	}, nil
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (fs *gcsFS) Remove(name string) error {
	obj := fs.client.Bucket(fs.bucket).Object(name)
	return obj.Delete(fs.ctx)
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (fs *gcsFS) RemoveAll(path string) error {
	// List all objects with the path prefix and delete them
	it := fs.client.Bucket(fs.bucket).Objects(fs.ctx, &storage.Query{
		Prefix: path,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		if err := fs.client.Bucket(fs.bucket).Object(attrs.Name).Delete(fs.ctx); err != nil {
			return err
		}
	}

	return nil
}

// Rename renames a file.
func (fs *gcsFS) Rename(oldname, newname string) error {
	// GCS doesn't support atomic rename, we need to copy and delete
	src := fs.client.Bucket(fs.bucket).Object(oldname)
	dst := fs.client.Bucket(fs.bucket).Object(newname)

	_, err := dst.CopierFrom(src).Run(fs.ctx)
	if err != nil {
		return err
	}

	return src.Delete(fs.ctx)
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (fs *gcsFS) Stat(name string) (os.FileInfo, error) {
	obj := fs.client.Bucket(fs.bucket).Object(name)
	attrs, err := obj.Attrs(fs.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, fmt.Errorf("file does not exist: %s", name)
		}
		return nil, err
	}

	return &gcsFileInfo{
		name:    name,
		size:    attrs.Size,
		modTime: attrs.Updated,
		isDir:   false,
	}, nil
}

// Name returns the name of the file system.
func (fs *gcsFS) Name() string {
	return "gcs"
}

// Chmod changes the mode of the named file to mode.
func (fs *gcsFS) Chmod(name string, mode os.FileMode) error {
	// GCS doesn't support file permissions
	return nil
}

// Chown changes the numeric uid and gid of the named file.
func (fs *gcsFS) Chown(name string, uid, gid int) error {
	// GCS doesn't support file ownership
	return nil
}

// Chtimes changes the access and modification times of the named file
func (fs *gcsFS) Chtimes(name string, atime, mtime time.Time) error {
	// GCS doesn't support changing times after creation
	return nil
}

// gcsFile implements afero.File for Google Cloud Storage
type gcsFile struct {
	fs     *gcsFS
	name   string
	mode   os.FileMode
	reader *storage.Reader
	writer *storage.Writer
	offset int64
}

// Close closes the file.
func (f *gcsFile) Close() error {
	if f.reader != nil {
		return f.reader.Close()
	}
	if f.writer != nil {
		return f.writer.Close()
	}
	return nil
}

// Read reads up to len(b) bytes from the File.
func (f *gcsFile) Read(b []byte) (n int, error) {
	if f.reader == nil {
		obj := f.fs.client.Bucket(f.fs.bucket).Object(f.name)
		reader, err := obj.NewReader(f.fs.ctx)
		if err != nil {
			return 0, err
		}
		f.reader = reader
	}
	
	return f.reader.Read(b)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off.
func (f *gcsFile) ReadAt(b []byte, off int64) (n int, error) {
	if f.reader == nil {
		obj := f.fs.client.Bucket(f.fs.bucket).Object(f.name)
		reader, err := obj.NewRangeReader(f.fs.ctx, off, int64(len(b)))
		if err != nil {
			return 0, err
		}
		defer reader.Close()
		return reader.Read(b)
	}
	
	return 0, fmt.Errorf("ReadAt not supported with active reader")
}

// Seek sets the offset for the next Read or Write.
func (f *gcsFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0: // Relative to origin
		f.offset = offset
	case 1: // Relative to current position
		f.offset += offset
	case 2: // Relative to end
		return 0, fmt.Errorf("seeking relative to end not supported")
	default:
		return 0, fmt.Errorf("invalid whence value")
	}
	
	return f.offset, nil
}

// Write writes len(b) bytes to the File.
func (f *gcsFile) Write(b []byte) (n int, error) {
	if f.writer == nil {
		obj := f.fs.client.Bucket(f.fs.bucket).Object(f.name)
		f.writer = obj.NewWriter(f.fs.ctx)
	}
	
	return f.writer.Write(b)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
func (f *gcsFile) WriteAt(b []byte, off int64) (n int, error) {
	return 0, fmt.Errorf("WriteAt not supported")
}

// Name returns the name of the file.
func (f *gcsFile) Name() string {
	return f.name
}

// Readdir reads directory entries.
func (f *gcsFile) Readdir(count int) ([]os.FileInfo, error) {
	var infos []os.FileInfo
	
	it := f.fs.client.Bucket(f.fs.bucket).Objects(f.fs.ctx, &storage.Query{
		Prefix: f.name,
	})

	for len(infos) < count || count <= 0 {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		info := &gcsFileInfo{
			name:    attrs.Name,
			size:    attrs.Size,
			modTime: attrs.Updated,
			isDir:   false,
		}
		infos = append(infos, info)

		if count > 0 && len(infos) >= count {
			break
		}
	}

	return infos, nil
}

// Readdirnames reads directory entry names.
func (f *gcsFile) Readdirnames(n int) ([]string, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}

	return names, nil
}

// Stat returns file info.
func (f *gcsFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(f.name)
}

// Sync commits the current contents of the file.
func (f *gcsFile) Sync() error {
	if f.writer != nil {
		return f.writer.Close()
	}
	return nil
}

// Truncate changes the size of the file.
func (f *gcsFile) Truncate(size int64) error {
	return fmt.Errorf("truncate not supported")
}

// WriteString writes a string to the file.
func (f *gcsFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

// gcsFileInfo implements os.FileInfo for Google Cloud Storage objects
type gcsFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

// Name returns the base name of the file.
func (fi *gcsFileInfo) Name() string {
	return fi.name
}

// Size returns the length in bytes for regular files.
func (fi *gcsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file mode bits.
func (fi *gcsFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}

// ModTime returns the modification time.
func (fi *gcsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the file is a directory.
func (fi *gcsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns the underlying data source.
func (fi *gcsFileInfo) Sys() interface{} {
	return nil
}