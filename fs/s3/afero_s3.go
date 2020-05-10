package afero_s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/stripslash"
	"github.com/spf13/afero"
	afero_s3 "github.com/wreulicke/afero-s3"
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

	return stripslash.NewStripSlashPathFs(afero_s3.NewFs(bucket, s3Int), 1), nil
}
