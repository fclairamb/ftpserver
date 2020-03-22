go get -t -v
gofmt -e -s -l -w .
goimports -l -w .
go vet
rem golint ./...
rem gocyclo -over 15 .
go test -race ./tests
