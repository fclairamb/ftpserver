// Package s3 provides an AWS S3 access layer
package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/fclairamb/afero-s3"
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

	conf := aws.Config{
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(access.Params["disable_ssl"] == "true"),
		S3ForcePathStyle: aws.Bool(access.Params["path_style"] == "true"),
	}

	if keyID != "" && secretAccessKey != "" {
		conf.Credentials = credentials.NewStaticCredentials(keyID, secretAccessKey, "")
	}

	if endpoint != "" {
		conf.Endpoint = aws.String(endpoint)
	}

	sess, errSession := session.NewSession(&conf)

	if errSession != nil {
		return nil, errSession
	}

	s3Fs := s3.NewFs(bucket, sess)

	// s3Fs = stripprefix.NewStripPrefixFs(s3Fs, 1)

	return s3Fs, nil
}
