package server

// This file is the driver part of the server. It must be implemented by anyone wanting to use the server.

type Driver interface {
	// Authenticate an user
	// Returns if the user could be authenticated
	CheckUser(user, pass string, userInfo *map[string]string) bool

	// List the files of a given directory
	// For each file, we have a map containing:
	// - name : The name of the file
	// - size : The size of the file
	GetFiles(userInfo *map[string]string) ([]map[string]string, error)
}
