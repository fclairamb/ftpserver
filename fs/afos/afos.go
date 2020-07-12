// Package afos provide an afero OS FS access layer
package afos

import (
	"errors"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ErrMissingBasePath is triggered when the basePath property isn't specified
var ErrMissingBasePath = errors.New("basePath must be specified")

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	basePath := access.Params["basePath"]
	if basePath == "" {
		return nil, ErrMissingBasePath
	}

	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}
