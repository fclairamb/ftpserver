module github.com/fclairamb/ftpserver

go 1.15

require (
	github.com/aws/aws-sdk-go v1.40.6
	github.com/fclairamb/afero-dropbox v0.1.0
	github.com/fclairamb/afero-gdrive v0.2.0
	github.com/fclairamb/afero-s3 v0.3.0
	github.com/fclairamb/ftpserverlib v0.14.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/pkg/sftp v1.13.2
	github.com/spf13/afero v1.6.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

// replace github.com/fclairamb/ftpserverlib => /Users/florent/go/src/github.com/fclairamb/ftpserverlib
// replace github.com/fclairamb/afero-s3 => /Users/florent/go/src/github.com/fclairamb/afero-s3
// replace github.com/fclairamb/afero-gdrive => /Users/florent/go/src/github.com/fclairamb/afero-gdrive
