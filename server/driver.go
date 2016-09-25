package server

// This file is the driver part of the server. It must be implemented by anyone wanting to use the server.

// Adding the ClientContext concept to be able to handle more than just UserInfo
type ClientContext interface {
	UserInfo() map[string]string
}

// Server driver
type Driver interface {
	// Authenticate an user
	// Returns if the user could be authenticated
	CheckUser(client ClientContext, user, pass string) error

	// Request to use a directory
	// Request access to user a directory
	GoToDirectory(client ClientContext, directory string) error

	// List the files of a given directory
	// For each file, we have a map containing:
	// - name : The name of the file
	// - size : The size of the file
	GetFiles(client ClientContext) ([]map[string]string, error)
}
