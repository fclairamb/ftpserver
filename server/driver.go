package server

import (
	"io"
	"os"
)

// This file is the driver part of the server. It must be implemented by anyone wanting to use the server.

// Server driver
type Driver interface {
	// Load some general settings around the server setup
	GetSettings() *Settings

	// When a user connects
	WelcomeUser(cc ClientContext) (string, error)

	// When a user disconnects
	UserLeft(cc ClientContext)

	// Authenticate an user
	// Return nil to accept the user
	CheckUser(cc ClientContext, user, pass string) error

	// Change current working directory
	ChangeDirectory(cc ClientContext, directory string) error

	// Create a directory
	MakeDirectory(cc ClientContext, directory string) error

	// List the files of the current working directory
	ListFiles(cc ClientContext) ([]os.FileInfo, error)

	// Upload a file
	OpenFile(cc ClientContext, path string, flag int) (FileContext, error)

	// Delete a file
	DeleteFile(cc ClientContext, path string) error

	// Get some info about a file
	GetFileInfo(cc ClientContext, path string) (os.FileInfo, error)

	// Move a file
	RenameFile(cc ClientContext, from, to string) error
}

// Adding the ClientContext concept to be able to handle more than just UserInfo
// Implemented by the server
type ClientContext interface {
	// Get current path
	Path() string

	// Custom value. This avoids having to create a mapping between the client.Id and our own internal system. We can
	// just store the driver's instance in the ClientContext
	MyInstance() interface{}

	// Set the custom value
	SetMyInstance(interface{})
}

// FileContext to use
type FileContext interface {
	io.Writer
	io.Reader
	io.Closer
	io.Seeker
}

// Server settings
type Settings struct {
	Host           string // Host to receive connections on
	Port           int    // Port to listen on
	MaxConnections int    // Max number of connections to accept
	MaxPassive     int    // Max number of passive connections per control connections to accept
	MonitorOn      bool   // To activate the monitor
	MonitorPort    int    // Port for the monitor to listen on
}
