module github.com/fclairamb/ftpserver

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.65
	github.com/fclairamb/afero-dropbox v0.1.0
	github.com/fclairamb/afero-gdrive v0.2.0
	github.com/fclairamb/afero-s3 v0.3.0
	github.com/fclairamb/ftpserverlib v0.13.2
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/pkg/sftp v1.13.1
	github.com/spf13/afero v1.6.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/oauth2 v0.0.0-20210622215436-a8dc77f794b6
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

// replace github.com/fclairamb/ftpserverlib => /Users/florent/go/src/github.com/fclairamb/ftpserverlib
// replace github.com/fclairamb/afero-s3 => /Users/florent/go/src/github.com/fclairamb/afero-s3
// replace github.com/fclairamb/afero-gdrive => /Users/florent/go/src/github.com/fclairamb/afero-gdrive
