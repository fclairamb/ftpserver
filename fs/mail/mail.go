// Package mail provides a mail access layer
package mail

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-mail/mail"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ErrNotImplemented is returned when something is not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrNotFound is returned when something is not found
var ErrNotFound = errors.New("not found")

// ErrInvalidParameter is returned when a parameter is invalid
var ErrInvalidParameter = errors.New("invalid parameter")

// Fs is a write-only afero.Fs implementation using mail as backend
type Fs struct {
	Dialer  mail.Dialer
	From    string
	To      string
	Subject string
	Message string
}

// File is the afero.File implementation
type File struct {
	Path    string
	Content []byte
	Fs      *Fs
	At      int64
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	port, err := strconv.Atoi(access.Params["Port"])
	if err != nil {
		return nil, err
	}

	if port < 1 || port > 65535 {
		port = 25
	}

	ssl, err2 := strconv.ParseBool(access.Params["SSL"])
	if err2 != nil {
		return nil, err2
	}

	var starttlspolicy mail.StartTLSPolicy

	switch access.Params["StartTLSPolicy"] {
	case "OpportunisticStartTLS":
		starttlspolicy = mail.OpportunisticStartTLS
	case "MandatoryStartTLS":
		starttlspolicy = mail.MandatoryStartTLS
	case "NoStartTLS":
		starttlspolicy = mail.NoStartTLS
	default:
		return nil, ErrInvalidParameter
	}

	f := &Fs{
		Dialer: mail.Dialer{
			Host:           access.Params["Host"],
			Port:           port,
			SSL:            ssl,
			StartTLSPolicy: starttlspolicy,
			Username:       access.Params["Username"],
			Password:       access.Params["Password"],
			LocalName:      access.Params["Localname"],
		},
		From:    access.Params["From"],
		To:      access.Params["To"],
		Subject: access.Params["Subject"],
		Message: access.Params["Message"],
	}

	return f, nil
}

// Name of the file
func (f *File) Name() string { return f.Path }

// Close closes the file transfer and does the mail sending
func (f *File) Close() error {
	if f.Fs == nil {
		return ErrNotFound
	}

	m := mail.NewMessage()
	m.SetHeader("From", f.Fs.From)
	m.SetHeader("To", f.Fs.To)
	m.SetHeader("Subject", f.Fs.Subject)
	m.SetBody("text/plain", fmt.Sprintf(f.Fs.Message, f.Path))
	m.AttachReader(f.Path, f)

	if err := f.Fs.Dialer.DialAndSend(m); err != nil {
		return err
	}

	f.Content = []byte{}
	f.At = 0

	return nil
}

// Read stores the received file content into the local buffer
func (f *File) Read(b []byte) (int, error) {
	n := 0

	if len(b) > 0 && int(f.At) == len(f.Content) {
		return 0, io.EOF
	}

	if len(f.Content)-int(f.At) >= len(b) {
		n = len(b)
	} else {
		n = len(f.Content) - int(f.At)
	}

	copy(b, f.Content[f.At:f.At+int64(n)])
	atomic.AddInt64(&f.At, int64(n))

	return n, nil
}

// ReadAt is not implemented
func (f *File) ReadAt(_ []byte, _ int64) (int, error) {
	return 0, ErrNotImplemented
}

// Truncate is not implemented
func (f *File) Truncate(_ int64) error {
	return nil
}

// Readdir is not implemented
func (f *File) Readdir(_ int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

// Readdirnames is not implemented
func (f *File) Readdirnames(_ int) ([]string, error) {
	return []string{}, nil
}

// Seek is not implemented
func (f *File) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

// Stat is not implemented
func (f *File) Stat() (os.FileInfo, error) {
	return nil, ErrNotImplemented
}

// Sync is not implemented
func (f *File) Sync() error {
	return nil
}

// WriteString is not implemented
func (f *File) WriteString(s string) (int, error) {
	return 0, ErrNotImplemented
}

// WriteAt is not implemented
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	return 0, ErrNotImplemented
}

func (f *File) Write(b []byte) (int, error) {
	f.Content = append(f.Content, b...)

	return len(b), nil
}

// Name of the filesystem
func (m *Fs) Name() string {
	return "mail"
}

// Chtimes is not implemented
func (m *Fs) Chtimes(name string, atime, mtime time.Time) error {
	return nil
}

// Chmod is not implemented
func (m *Fs) Chmod(name string, mode os.FileMode) error {
	return nil
}

// Rename is not implemented
func (m *Fs) Rename(name string, newname string) error {
	return nil
}

// Chown is not implemented
func (m *Fs) Chown(string, int, int) error {
	return nil
}

// RemoveAll is not implemented
func (m *Fs) RemoveAll(name string) error {
	return nil
}

// Remove is not implemented
func (m *Fs) Remove(name string) error {
	return nil
}

// Mkdir is not implemented
func (m *Fs) Mkdir(name string, mode os.FileMode) error {
	return nil
}

// MkdirAll is not implemented
func (m *Fs) MkdirAll(name string, mode os.FileMode) error {
	return nil
}

// Open opens a file buffer
func (m *Fs) Open(name string) (afero.File, error) {
	return &File{Path: name, Fs: m}, nil
}

// Create creates a file buffer
func (m *Fs) Create(name string) (afero.File, error) {
	return &File{Path: name, Fs: m}, nil
}

// OpenFile opens a file buffer
func (m *Fs) OpenFile(name string, flag int, mode os.FileMode) (afero.File, error) {
	return &File{Path: name, Fs: m}, nil
}

// Stat is not implemented
func (m *Fs) Stat(name string) (os.FileInfo, error) {
	return nil, &os.PathError{Op: "stat", Path: name, Err: nil}
}

// LstatIfPossible is not implemented
func (m *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	return nil, false, &os.PathError{Op: "lstat", Path: name, Err: nil}
}
