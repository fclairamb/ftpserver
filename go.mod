module github.com/fclairamb/ftpserver

go 1.14

require (
	github.com/aws/aws-sdk-go v1.35.33
	github.com/fclairamb/afero-gdrive v0.0.0-20201202000508-699f74d5f307
	github.com/aws/aws-sdk-go v1.36.2
	github.com/fclairamb/afero-s3 v0.1.1
	github.com/fclairamb/ftpserverlib v0.9.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/pkg/sftp v1.12.0
	github.com/spf13/afero v1.4.1
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/oauth2 v0.0.0-20201207163604-931764155e3f
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

// replace github.com/fclairamb/ftpserverlib => /Users/florent/go/src/github.com/fclairamb/ftpserverlib
// replace github.com/fclairamb/afero-s3 => /Users/florent/go/src/github.com/fclairamb/afero-s3
// replace github.com/fclairamb/afero-gdrive => /Users/florent/go/src/github.com/fclairamb/afero-gdrive
