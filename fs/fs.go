// Package fs provides all the core features related to file-system access
package fs

import (
	"fmt"
	"io"
	"log/slog"

	snd "github.com/fclairamb/afero-snd"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/afos"
	"github.com/fclairamb/ftpserver/fs/dropbox"
	"github.com/fclairamb/ftpserver/fs/gcs"
	"github.com/fclairamb/ftpserver/fs/gdrive"
	"github.com/fclairamb/ftpserver/fs/keycloak"
	"github.com/fclairamb/ftpserver/fs/mail"
	"github.com/fclairamb/ftpserver/fs/s3"
	"github.com/fclairamb/ftpserver/fs/sftp"
	"github.com/fclairamb/ftpserver/fs/telegram"
)

// UnsupportedFsError is returned when the described file system is not supported
type UnsupportedFsError struct {
	Type string
}

func (err UnsupportedFsError) Error() string {
	return fmt.Sprintf("Unsupported FS: %s", err.Type)
}

// LoadFs loads a file system from an access description.
//
// The returned closer releases what the backend holds (an SFTP connection, a Telegram poller). It is nil when the
// backend holds nothing. The wrappers applied on top of the backend (read-only, sync & delete) hide its Close
// method, so the caller must keep this closer and call it when the file system is no longer used.
func LoadFs(access *confpar.Access, logger *slog.Logger) (afero.Fs, io.Closer, error) {
	var fs afero.Fs
	var err error

	switch access.Fs {
	case "os":
		fs, err = afos.LoadFs(access)
	case "s3":
		fs, err = s3.LoadFs(access)
	case "gcs":
		fs, err = gcs.LoadFs(access)
	case "sftp":
		fs, err = sftp.LoadFs(access, logger.With("component", "sftp"))
	case "mail":
		fs, err = mail.LoadFs(access)
	case "gdrive":
		fs, err = gdrive.LoadFs(access, logger.With("component", "gdrive"))
	case "keycloak":
		fs, err = keycloak.LoadFs(access)
	case "dropbox":
		fs, err = dropbox.LoadFs(access)
	case "telegram":
		fs, err = telegram.LoadFs(access, logger.With("component", "telegram"))
	default:
		fs, err = nil, &UnsupportedFsError{Type: access.Fs}
	}

	if err != nil {
		return nil, nil, err
	}

	closer, _ := fs.(io.Closer)

	if access.ReadOnly {
		fs = afero.NewReadOnlyFs(fs)
	}

	// If we're defining a dubious behavior, we can use it
	if access.SyncAndDelete != nil && access.SyncAndDelete.Enable {
		var temp afero.Fs

		if access.SyncAndDelete.Directory != "" {
			temp = afero.NewBasePathFs(afero.NewOsFs(), access.SyncAndDelete.Directory)
		}

		fs, err = snd.NewFs(&snd.Config{
			Destination: fs,
			Temporary:   temp,
			Logger:      nil,
		})
	}

	if err != nil {
		if closer != nil {
			_ = closer.Close()
		}

		return nil, nil, err
	}

	return fs, closer, nil
}
