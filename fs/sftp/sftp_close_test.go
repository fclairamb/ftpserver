package sftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// startSFTPServer starts an in-process SSH server with an SFTP subsystem. It returns its address, its host key and
// a channel that receives a value each time a client connection ends.
func startSFTPServer(t *testing.T) (string, ssh.PublicKey, <-chan struct{}) {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("cannot generate key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("cannot create signer: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) != "secret" {
				return nil, ssh.ErrNoAuth
			}

			return nil, nil
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	closed := make(chan struct{}, 10)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go serveSFTP(conn, config, closed)
		}
	}()

	return listener.Addr().String(), signer.PublicKey(), closed
}

func serveSFTP(conn net.Conn, config *ssh.ServerConfig, closed chan<- struct{}) {
	defer func() { closed <- struct{}{} }()

	serverConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}

	go ssh.DiscardRequests(reqs)

	go func() {
		for newChan := range chans {
			if newChan.ChannelType() != "session" {
				_ = newChan.Reject(ssh.UnknownChannelType, "unsupported")

				continue
			}

			channel, requests, err := newChan.Accept()
			if err != nil {
				continue
			}

			go func() {
				for req := range requests {
					ok := req.Type == "subsystem" && len(req.Payload) > 4 && string(req.Payload[4:]) == "sftp"
					_ = req.Reply(ok, nil)

					if ok {
						server, err := sftp.NewServer(channel)
						if err != nil {
							return
						}

						_ = server.Serve()
						_ = server.Close()
					}
				}
			}()
		}
	}()

	// Wait returns when the client closes the connection
	_ = serverConn.Wait()
}

// TestCloseClosesConnection makes sure closing the file system closes its SSH connection
func TestCloseClosesConnection(t *testing.T) {
	addr, hostKey, closed := startSFTPServer(t)

	fs, err := LoadFs(&confpar.Access{
		Fs: "sftp",
		Params: map[string]string{
			"hostname": addr,
			"username": "user",
			"password": "secret",
			"host_key": string(ssh.MarshalAuthorizedKey(hostKey)),
		},
	}, nil)
	if err != nil {
		t.Fatalf("LoadFs: %v", err)
	}

	if _, err := fs.Stat("/"); err != nil {
		t.Fatalf("Stat: %v", err)
	}

	closer, ok := fs.(io.Closer)
	if !ok {
		t.Fatalf("the SFTP file system %T should implement io.Closer", fs)
	}

	select {
	case <-closed:
		t.Fatal("the connection should still be open before Close")
	case <-time.After(100 * time.Millisecond):
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the server should see the connection closed after Close")
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
