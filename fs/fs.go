// Package fs provides all the core features related to file-system access
package fs

import (
	"fmt"

	"github.com/fclairamb/ftpserverlib/log"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/afos"
	"github.com/fclairamb/ftpserver/fs/dropbox"
	"github.com/fclairamb/ftpserver/fs/gdrive"
	"github.com/fclairamb/ftpserver/fs/mail"
	"github.com/fclairamb/ftpserver/fs/s3"
	"github.com/fclairamb/ftpserver/fs/sftp"
)

// UnsupportedFsError is returned when the described file system is not supported
type UnsupportedFsError struct {
	error
	Type string
}

func (err UnsupportedFsError) Error() string {
	return fmt.Sprintf("Unsupported FS: %s", err.Type)
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access, logger log.Logger) (afero.Fs, error) {
	var fs afero.Fs
	var err error

	switch access.Fs {
	case "os":
		fs, err = afos.LoadFs(access)
	case "s3":
		fs, err = s3.LoadFs(access)
	case "sftp":
		fs, err = sftp.LoadFs(access)
	case "mail":
		fs, err = mail.LoadFs(access)
	case "gdrive":
		fs, err = gdrive.LoadFs(access, logger)
	case "dropbox":
		fs, err = dropbox.LoadFs(access)
	default:
		fs, err = nil, &UnsupportedFsError{Type: access.Fs}
	}

	if err != nil && access.ReadOnly {
		fs = afero.NewReadOnlyFs(fs)
	}

	return fs, err
}
