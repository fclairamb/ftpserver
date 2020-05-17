package afero_s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	afero_s3 "github.com/fclairamb/afero-s3"
	// "github.com/chonthu/aferoS3"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/spf13/afero"
)

func LoadFs(access *confpar.Access) (afero.Fs, error) {
	region := access.Params["region"]
	bucket := access.Params["bucket"]
	keyId := access.Params["access_key_id"]
	secretAccessKey := access.Params["secret_access_key"]

	sess, errSession := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(keyId, secretAccessKey, ""),
	})

	if errSession != nil {
		return nil, errSession
	}

	s3Int := s3.New(sess)
	s3Fs := afero_s3.NewFs(bucket, s3Int)
	// withoutTrailingSlash := stripslash.NewStripSlashPathFs(s3Fs, 1)
	return s3Fs, nil
	// return aferoS3.NewS3Fs(sess, bucket), nil
}
