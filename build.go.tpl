package main

var (
	// BuildVersion is the current version of the program
	BuildVersion = "{{ .Version }}"

	// BuildDate is the time the program was built
	BuildDate = "{{ .BuildDate }}"

	// Commit is the git hash of the program
	Commit = "{{ .GitCommitHashShort }}"
)
