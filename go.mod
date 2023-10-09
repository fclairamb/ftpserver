module github.com/clicknclear/ftpserver

go 1.19

require (
	github.com/aws/aws-sdk-go v1.45.24
	github.com/fclairamb/afero-dropbox v0.1.0
	github.com/fclairamb/afero-gdrive v0.3.0
	github.com/fclairamb/afero-snd v0.1.0
	github.com/clicknclear/afero-s3 v0.4.3
	github.com/fclairamb/ftpserverlib v0.22.0
	github.com/fclairamb/go-log v0.4.1
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/pkg/sftp v1.13.6
	github.com/spf13/afero v1.10.0
	github.com/tidwall/sjson v1.2.5
	golang.org/x/crypto v0.14.0
	golang.org/x/oauth2 v0.13.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	github.com/dropbox/dropbox-sdk-go-unofficial v5.6.0+incompatible // indirect
	github.com/go-kit/log v0.2.1
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.16.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/api v0.126.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

// replace github.com/fclairamb/ftpserverlib => ../ftpserverlib
// replace github.com/clicknclear/afero-s3 => ../afero-s3

// replace github.com/fclairamb/afero-gdrive => /Users/florent/go/src/github.com/fclairamb/afero-gdrive
// replace github.com/fclairamb/afero-snd => /Users/florent/go/src/github.com/fclairamb/afero-snd
