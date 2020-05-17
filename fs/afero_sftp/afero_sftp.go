package afero_sftp

import (
	"fmt"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"github.com/spf13/afero/sftpfs"
	"golang.org/x/crypto/ssh"
	"net"
)

func LoadFs(access *confpar.Access) (afero.Fs, error) {
	par := access.Params
	config := &ssh.ClientConfig{
		User: par["usename"],
		Auth: []ssh.AuthMethod{
			ssh.Password(par["password"]),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// Dial your ssh server.
	conn, errSsh := ssh.Dial("tcp", par["hostname"], config)
	if errSsh != nil {
		return nil, fmt.Errorf("unable to connect: %s", errSsh)
	}

	client, errSftp := sftp.NewClient(conn)
	if errSftp != nil {
		return nil, fmt.Errorf("unable to setup sftp: %s", errSftp)
	}

	return sftpfs.New(client), nil
}
