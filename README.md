# Golang FTP Server library

The goal is to provide a comprehensive FTP Server library.

Note: this is a fork of [andrewarrow/paradise_ftp](https://github.com/andrewarrow/paradise_ftp) but many things have been changed since then.

## Current status of the project

### Features

 * Uploading and downloading files
 * Directory listing
 * File and directory deletion and renaming
 * TLS support
 * Complete driver for all the above features
 * Passive socket connections (EPSV and PASV commands, no active connections)
 * Small memory footprint
 * graceful restarts by sending kill -USR2 pid

## The driver

### The API
```go
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
	OpenFile(cc ClientContext, path string, flag int) (FileStream, error)

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

	// SetDebug activates the debugging of this connection commands
	SetDebug(debug bool)

	// Debug returns the current debugging status of this connection commands
	Debug() bool
}

// FileStream is a read or write closeable stream
type FileStream interface {
	io.Writer
	io.Reader
	io.Closer
	io.Seeker // <-- Will be used for "REST" command
}

// Settings define all the server settings
type Settings struct {
	Host           string // Host to receive connections on
	Port           int    // Port to listen on
	MaxConnections int    // Max number of connections to accept
	MonitorOn      bool   // To activate the monitor
	MonitorPort    int    // Port for the monitor to listen on
}
```

### Sample implementation

Have a look at the [sample driver](https://github.com/fclairamb/ftpserver/tree/master/sample). It shows how you can plug your FTP server to your file system.


## Sample run
```
$ ftp ftp://a:a@localhost:2121
Trying ::1...
Connected to localhost.
220 Welcome on https://github.com/fclairamb/ftpserver
331 OK
230 Password ok, continue
Remote system type is UNIX.
Using binary mode to transfer files.
200 Type set to binary
ftp> put iMX7D_RM_Rev_B.pdf 
local: iMX7D_RM_Rev_B.pdf remote: iMX7D_RM_Rev_B.pdf
229 Entering Extended Passive Mode (|||62362|)
150 Using transfer connection
100% |******************************************************************************************************************************************************************| 44333 KiB  635.92 MiB/s    00:00 ETA
226 OK, received 45397173 bytes
45397173 bytes sent in 00:00 (538.68 MiB/s)
ftp> cd virtual
250 CD worked on /virtual
ftp> ls
229 Entering Extended Passive Mode (|||62369|)
150 Using transfer connection
-rw-rw-rw- 1 ftp ftp         1024 Sep 28 01:44 localpath.txt
-rw-rw-rw- 1 ftp ftp         2048 Sep 28 01:44 file2.txt

226 Closing data connection, sent some bytes
ftp> get localpath.txt
local: localpath.txt remote: localpath.txt
229 Entering Extended Passive Mode (|||62371|)
150 Using transfer connection
    67      241.43 KiB/s 
226 OK, sent 67 bytes
67 bytes received in 00:00 (160.36 KiB/s)
ftp> ^D
221 Goodbye
$ more localpath.txt 
/var/folders/vk/vgsfkf9975xfrc4_fk102g200000gn/T/ftpserver020090599
$ shasum /var/folders/vk/vgsfkf9975xfrc4_fk102g200000gn/T/ftpserver020090599/iMX7D_RM_Rev_B.pdf 
03b3686b31867fb14d3f3a61e20d28a029883a32  /var/folders/vk/vgsfkf9975xfrc4_fk102g200000gn/T/ftpserver020090599/iMX7D_RM_Rev_B.pdf
$ more localpath.txt 
$ shasum iMX7D_RM_Rev_B.pdf 
03b3686b31867fb14d3f3a61e20d28a029883a32  iMX7D_RM_Rev_B.pdf
```

## History of the project

I wanted to make a system which would accept files through FTP and redirect them to something else. Go seemed like the obvious choice and there seemed to be a lot of libraries available but it turns out none of them were in a useable state.

* [micahhausler/go-ftp](https://github.com/micahhausler/go-ftp) is a  minimalistic implementation 
* [shenfeng/ftpd.go](https://github.com/shenfeng/ftpd.go) is very basic and 4 years old.
* [yob/graval](https://github.com/yob/graval) is 3 years old and “experimental”.
* [goftp/server](https://github.com/goftp/server) seemed OK but I couldn't use it both Filezilla and the MacOs ftp client.
* [andrewarrow/paradise_ftp](https://github.com/andrewarrow/paradise_ftp) - Was the only one of the list I could test right away. But it turns out there's many features missing and I found that there was many things I could improve.

That's why I forked from this last one.
