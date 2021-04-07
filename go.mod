module github.com/fclairamb/ftpserver

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.15
	github.com/fclairamb/afero-dropbox v0.1.0
	github.com/fclairamb/afero-gdrive v0.2.0
	github.com/fclairamb/afero-s3 v0.3.0
	github.com/fclairamb/ftpserverlib v0.13.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/pkg/sftp v1.13.0
	github.com/spf13/afero v1.6.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

// replace github.com/fclairamb/ftpserverlib => /Users/florent/go/src/github.com/fclairamb/ftpserverlib
// replace github.com/fclairamb/afero-s3 => /Users/florent/go/src/github.com/fclairamb/afero-s3
// replace github.com/fclairamb/afero-gdrive => /Users/florent/go/src/github.com/fclairamb/afero-gdrive
