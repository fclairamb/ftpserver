package server

// This file is the driver part of the server. It must be implemented by anyone wanting to use the server.

// Adding the ClientContext concept to be able to handle more than just UserInfo
type ClientContext interface {
	// Get userInfo
	UserInfo() map[string]string

	// Get current path
	Path() string
}

// Server driver
type Driver interface {
	// Load some general settings around the server setup
	GetSettings() *Settings

	// Welcome a user
	WelcomeUser(cc ClientContext) (string, error)

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

	// Called when a user disconnects
	UserLeft(cc ClientContext)
}

// Settings are part of the driver
type Settings struct {
	Host           string // Host to receive connections on
	Port           int    // Port to listen on
	MaxConnections int    // Max number of connections to accept
	MaxPassive     int    // Max number of passive connections per control connections to accept
	MonitorOn      bool   // To activate the monitor
	MonitorPort    int    // Port for the monitor to listen on
	Exec           string
}
