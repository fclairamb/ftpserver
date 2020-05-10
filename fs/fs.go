package fs

import (
	"fmt"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/afero_os"
	afero_s3 "github.com/fclairamb/ftpserver/fs/s3"
	"github.com/spf13/afero"
)

func LoadFs(access *confpar.Access) (afero.Fs, error) {
	switch access.Fs {
	case "os":
		return afero_os.LoadFs(access)
	case "s3":
		return afero_s3.LoadFs(access)
	default:
		return nil, fmt.Errorf("Fs not supported: %s", access.Fs)
	}
}
