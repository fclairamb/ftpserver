// Package sftp provides an SFTP connection layer
package sftp

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"github.com/spf13/afero/sftpfs"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ConnectionError is returned when a connection occurs while connecting to the SFTP server
type ConnectionError struct {
	Source error
}

func (err ConnectionError) Error() string {
	return fmt.Sprintf("Could not connect to SFTP host: %#v", err.Source)
}

// ErrNoHostKeyVerification is returned when no host key verification method has been configured.
// Verifying the host key protects against man-in-the-middle attacks, so it must either be set up
// (with "known_hosts" or "host_key") or explicitly disabled (with "insecure_ignore_host_key").
var ErrNoHostKeyVerification = errors.New(
	`sftp: no host key verification configured: set "known_hosts" or "host_key", ` +
		`or set "insecure_ignore_host_key" to "true" to disable verification (unsafe)`)

// hostKeyCallback builds the SSH host key verification callback from the access parameters.
//
// It supports, in order of precedence:
//   - insecure_ignore_host_key="true": accepts any host key (vulnerable to MITM, opt-in only)
//   - host_key: pins a single expected public key (authorized_keys / known_hosts line format)
//   - known_hosts: validates the host key against an OpenSSH known_hosts file
func hostKeyCallback(par map[string]string, logger *slog.Logger) (ssh.HostKeyCallback, error) {
	if par["insecure_ignore_host_key"] == "true" {
		if logger != nil {
			logger.Warn("SFTP host key verification is disabled, " +
				"the connection is vulnerable to man-in-the-middle attacks")
		}

		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // explicitly opted-in by the user
	}

	if hostKey := par["host_key"]; hostKey != "" {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
		if err != nil {
			return nil, fmt.Errorf(`sftp: invalid "host_key" parameter: %w`, err)
		}

		return ssh.FixedHostKey(key), nil
	}

	if knownHostsFile := par["known_hosts"]; knownHostsFile != "" {
		callback, err := knownhosts.New(knownHostsFile)
		if err != nil {
			return nil, fmt.Errorf(`sftp: cannot load "known_hosts" file: %w`, err)
		}

		return callback, nil
	}

	return nil, ErrNoHostKeyVerification
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access, logger *slog.Logger) (afero.Fs, error) {
	par := access.Params

	hostKeyCB, err := hostKeyCallback(par, logger)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: par["username"],
		Auth: []ssh.AuthMethod{
			ssh.Password(par["password"]),
		},
		HostKeyCallback: hostKeyCB,
	}

	// Dial the ssh server.
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
