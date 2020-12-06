// Package fs provides all the core features related to file-system access
package fs

import (
	"fmt"

	"github.com/fclairamb/ftpserverlib/log"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/afos"
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
	switch access.Fs {
	case "os":
		return afos.LoadFs(access)
	case "s3":
		return s3.LoadFs(access)
	case "sftp":
		return sftp.LoadFs(access)
	case "mail":
		return mail.LoadFs(access)
	case "gdrive":
		return gdrive.LoadFs(access, logger)
	default:
		return nil, &UnsupportedFsError{Type: access.Fs}
	}
}
