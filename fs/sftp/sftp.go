// Package sftp provides an SFTP connection layer
package sftp

import (
	"fmt"
	"net"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"github.com/spf13/afero/sftpfs"
	"golang.org/x/crypto/ssh"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ConnectionError is returned when a connection occurs while connecting to the SFTP server
type ConnectionError struct {
	error
	Source error
}

func (err ConnectionError) Error() string {
	return fmt.Sprintf("Could not connect to SFTP host: %#v", err.Source)
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	par := access.Params
	config := &ssh.ClientConfig{
		User: par["username"],
		Auth: []ssh.AuthMethod{
			ssh.Password(par["password"]),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// Dial your ssh server.
	conn, errSSH := ssh.Dial("tcp", par["hostname"], config)
	if errSSH != nil {
		return nil, &ConnectionError{Source: errSSH}
	}

	client, errSftp := sftp.NewClient(conn)
	if errSftp != nil {
		return nil, &ConnectionError{Source: errSftp}
	}

	return sftpfs.New(client), nil
}
