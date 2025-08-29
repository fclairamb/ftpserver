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

	return gcsfs.NewFs(ctx, client, bucket), nil
}

