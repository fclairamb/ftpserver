// Package gcs provides a Google Cloud Storage access layer
package gcs

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/spf13/afero"
	"github.com/spf13/afero/gcsfs"
	"google.golang.org/api/option"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ErrMissingBucket is returned when the bucket parameter isn't specified
var ErrMissingBucket = errors.New("bucket parameter is required for GCS")

// LoadFs loads a GCS file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	bucket := access.Params["bucket"]
	if bucket == "" {
		return nil, ErrMissingBucket
	}

	ctx := context.Background()

	var opts []option.ClientOption

	if keyFile := access.Params["key_file"]; keyFile != "" {
		// Use a service account key file. Without it, the application default credentials are used.
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, keyFile))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	gcsFs, err := gcsfs.NewGcsFSFromClient(ctx, client)
	if err != nil {
		return nil, err
	}

	// gcsfs uses the first path component as the bucket name, each access is restricted to its bucket
	return afero.NewBasePathFs(gcsFs, "/"+bucket), nil
}
