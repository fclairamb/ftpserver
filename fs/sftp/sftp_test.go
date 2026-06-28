package sftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const testAddr = "192.168.168.11:22"

func testRemote(t *testing.T) net.Addr {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", testAddr)
	if err != nil {
		t.Fatalf("cannot resolve test address: %v", err)
	}

	return addr
}

func newHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("cannot generate key: %v", err)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("cannot wrap key: %v", err)
	}

	return sshPub
}

func assertHostKey(t *testing.T, callback ssh.HostKeyCallback, accepted, rejected ssh.PublicKey) {
	t.Helper()

	if err := callback(testAddr, testRemote(t), accepted); err != nil {
		t.Fatalf("expected host key to be accepted, got %v", err)
	}

	if err := callback(testAddr, testRemote(t), rejected); err == nil {
		t.Fatal("expected host key to be rejected")
	}
}

func TestHostKeyCallbackRequiresConfiguration(t *testing.T) {
	_, err := hostKeyCallback(map[string]string{}, nil)
	if !errors.Is(err, ErrNoHostKeyVerification) {
		t.Fatalf("expected ErrNoHostKeyVerification, got %v", err)
	}
}

func TestHostKeyCallbackInsecure(t *testing.T) {
	callback, err := hostKeyCallback(map[string]string{"insecure_ignore_host_key": "true"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cbErr := callback(testAddr, testRemote(t), newHostKey(t)); cbErr != nil {
		t.Fatalf("insecure callback should accept any key, got %v", cbErr)
	}
}

func TestHostKeyCallbackFixed(t *testing.T) {
	expected := newHostKey(t)

	callback, err := hostKeyCallback(
		map[string]string{"host_key": string(ssh.MarshalAuthorizedKey(expected))},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHostKey(t, callback, expected, newHostKey(t))
}

func TestHostKeyCallbackFixedInvalid(t *testing.T) {
	_, err := hostKeyCallback(map[string]string{"host_key": "this is not a key"}, nil)
	if err == nil {
		t.Fatal("expected an error for an invalid host_key")
	}
}

func TestHostKeyCallbackKnownHosts(t *testing.T) {
	expected := newHostKey(t)

	file, err := os.CreateTemp(t.TempDir(), "known_hosts")
	if err != nil {
		t.Fatalf("cannot create known_hosts file: %v", err)
	}

	if _, err = file.WriteString(knownhosts.Line([]string{testAddr}, expected) + "\n"); err != nil {
		t.Fatalf("cannot write known_hosts file: %v", err)
	}

	if err = file.Close(); err != nil {
		t.Fatalf("cannot close known_hosts file: %v", err)
	}

	callback, err := hostKeyCallback(map[string]string{"known_hosts": file.Name()}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHostKey(t, callback, expected, newHostKey(t))
}

func TestHostKeyCallbackKnownHostsMissingFile(t *testing.T) {
	_, err := hostKeyCallback(map[string]string{"known_hosts": "/non/existent/known_hosts"}, nil)
	if err == nil {
		t.Fatal("expected an error for a missing known_hosts file")
	}
}
