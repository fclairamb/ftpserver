// Package afos provide an afero OS FS access layer
package afos

import (
	"errors"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	basePath := access.Params["basePath"]
	if basePath == "" {
		return nil, errors.New("basePath must be specified")
	}

	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}
