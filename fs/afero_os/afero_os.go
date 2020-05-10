package afero_os

import (
	"errors"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/spf13/afero"
)

func LoadFs(access *confpar.Access) (afero.Fs, error) {
	basePath := access.Params["basePath"]
	if basePath == "" {
		return nil, errors.New("basePath must be specified")
	}
	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}
