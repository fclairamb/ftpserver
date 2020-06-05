module github.com/fclairamb/ftpserver

go 1.14

require (
	github.com/aws/aws-sdk-go v1.31.10
	github.com/fclairamb/afero-s3 v0.1.0
	github.com/fclairamb/ftpserverlib v0.7.0
	github.com/go-kit/kit v0.10.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/pkg/sftp v1.11.0
	github.com/spf13/afero v1.2.2
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550
)

// replace github.com/fclairamb/ftpserverlib => /Users/florent/go/src/github.com/fclairamb/ftpserverlib
// replace github.com/fclairamb/afero-s3 => /Users/florent/go/src/github.com/fclairamb/afero-s3
