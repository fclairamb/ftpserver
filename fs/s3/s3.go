// Package s3 provides an AWS S3 access layer
package s3

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	aferos3 "github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ErrChmodNotSupported is returned by Chmod when the access doesn't set the "chmod_acl" param
var ErrChmodNotSupported = errors.New("chmod is not supported on s3")

// noChmodFs refuses Chmod and passes every other call to the wrapped file system
type noChmodFs struct {
	afero.Fs
}

// Chmod refuses to change the mode, without calling the backend
func (f *noChmodFs) Chmod(name string, _ os.FileMode) error {
	return &os.PathError{Op: "chmod", Path: name, Err: ErrChmodNotSupported}
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	endpoint := access.Params["endpoint"]
	region := access.Params["region"]
	bucket := access.Params["bucket"]
	keyID := access.Params["access_key_id"]
	secretAccessKey := access.Params["secret_access_key"]
	disableSSL := access.Params["disable_ssl"] == "true"
	pathStyle := access.Params["path_style"] == "true"
	basePath := access.Params["basePath"]
	chmodACL := access.Params["chmod_acl"] == "true"

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
		endpointURL := endpointWithScheme(endpoint, disableSSL)
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

	var fs afero.Fs = s3Fs
	if !chmodACL {
		// afero-s3 maps a chmod mode to a canned ACL (644 -> public-read), which would let any FTP user
		// with write access make objects public through SITE CHMOD. It is refused unless the operator opts in.
		fs = &noChmodFs{Fs: fs}
	}

	if basePath != "" {
		// Strip leading/trailing slashes: S3 object keys must not start with "/"
		// and a trailing slash would produce double slashes in key names.
		basePath = strings.Trim(basePath, "/")
		fs = afero.NewBasePathFs(fs, basePath)
	}

	return fs, nil
}

// endpointWithScheme adds the URL scheme to an endpoint that doesn't have one: "http" if SSL is disabled,
// "https" otherwise. An endpoint that already has a scheme is returned as is.
func endpointWithScheme(endpoint string, disableSSL bool) string {
	if strings.Contains(endpoint, "://") {
		return endpoint
	}

	if disableSSL {
		return "http://" + endpoint
	}

	return "https://" + endpoint
}
