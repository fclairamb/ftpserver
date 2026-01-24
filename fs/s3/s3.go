// Package s3 provides an AWS S3 access layer
package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	aferos3 "github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	endpoint := access.Params["endpoint"]
	region := access.Params["region"]
	bucket := access.Params["bucket"]
	keyID := access.Params["access_key_id"]
	secretAccessKey := access.Params["secret_access_key"]
	disableSSL := access.Params["disable_ssl"] == "true"
	pathStyle := access.Params["path_style"] == "true"

	// Build config options for AWS SDK v2
	var configOpts []func(*config.LoadOptions) error

	if region != "" {
		configOpts = append(configOpts, config.WithRegion(region))
	}

	if keyID != "" && secretAccessKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keyID, secretAccessKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), configOpts...)
	if err != nil {
		return nil, err
	}

	// Build S3 client options
	var s3Opts []func(*s3.Options)

	if endpoint != "" {
		endpointURL := endpoint
		if disableSSL {
			// Ensure HTTP scheme if SSL is disabled
			if len(endpointURL) > 0 && endpointURL[0] != 'h' {
				endpointURL = "http://" + endpointURL
			}
		}
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	}

	if pathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	// Create S3 client with options
	client := s3.NewFromConfig(cfg, s3Opts...)

	s3Fs := aferos3.NewFsFromClient(bucket, client)

	return s3Fs, nil
}
