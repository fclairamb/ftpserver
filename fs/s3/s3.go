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
	region := access.Params["region"]
	bucket := access.Params["bucket"]
	keyID := access.Params["access_key_id"]
	secretAccessKey := access.Params["secret_access_key"]

	sess, errSession := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(keyID, secretAccessKey, ""),
	})

	if errSession != nil {
		return nil, errSession
	}

	s3Fs := s3.NewFs(bucket, sess)

	// s3Fs = stripprefix.NewStripPrefixFs(s3Fs, 1)

	return s3Fs, nil
}
