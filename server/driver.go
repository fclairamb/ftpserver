package server

import (
	"io"
	"os"
)

// This file is the driver part of the server. It must be implemented by anyone wanting to use the server.

// ServerDriver handles the authentication and ClientHandlingDriver selection
type ServerDriver interface {
	// Load some general settings around the server setup
	GetSettings() *Settings

	// WelcomeUser is called to send the very first welcome message
	WelcomeUser(cc ClientContext) (string, error)

	// UserLeft is called when the user disconnects, even if he never authenticated
	UserLeft(cc ClientContext)

	// AuthUser authenticates the user and selects an handling driver
	AuthUser(cc ClientContext, user, pass string) (ClientHandlingDriver, error)
}

// ClientHandlingDriver handles the file system access logic
type ClientHandlingDriver interface {
	// ChangeDirectory changes the current working directory
	ChangeDirectory(cc ClientContext, directory string) error

	// MakeDirectory creates a directory
	MakeDirectory(cc ClientContext, directory string) error

	// ListFiles lists the files around a directory
	ListFiles(cc ClientContext) ([]os.FileInfo, error)

	// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
	OpenFile(cc ClientContext, path string, flag int) (FileContext, error)

	// DeleteFile deletes a file or a directory
	DeleteFile(cc ClientContext, path string) error

	// GetFileInfo gets some info around a file or a directory
	GetFileInfo(cc ClientContext, path string) (os.FileInfo, error)

	// RenameFile renames a file or a directory
	RenameFile(cc ClientContext, from, to string) error
}

// ClientContext is implemented on the server side to provide some access to few data around the client
type ClientContext interface {
	// Get current path
	Path() string

	// Custom value. This avoids having to create a mapping between the client.Id and our own internal system. We can
	// just store the driver's instance in the ClientContext
	MyInstance() interface{}

	// Set the custom value
	SetMyInstance(interface{})
}

// FileContext is read or write closeable stream
type FileContext interface {
	io.Writer
	io.Reader
	io.Closer
	// io.Seeker <-- Not use at that stage
}

// Settings define all the server settings
type Settings struct {
	Host           string // Host to receive connections on
	Port           int    // Port to listen on
	MaxConnections int    // Max number of connections to accept
	MaxPassive     int    // Max number of passive connections per control connections to accept
	MonitorOn      bool   // To activate the monitor
	MonitorPort    int    // Port for the monitor to listen on
}
