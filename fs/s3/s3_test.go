package s3

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
)

func TestEndpointWithScheme(t *testing.T) {
	for _, tc := range []struct {
		endpoint   string
		disableSSL bool
		want       string
	}{
		{"minio:9000", true, "http://minio:9000"},
		{"minio:9000", false, "https://minio:9000"},
		// Hostnames starting with an "h" used to be considered as already having a scheme
		{"host.example.com:9000", true, "http://host.example.com:9000"},
		{"http://minio:9000", false, "http://minio:9000"},
		{"https://s3.amazonaws.com", true, "https://s3.amazonaws.com"},
	} {
		if got := endpointWithScheme(tc.endpoint, tc.disableSSL); got != tc.want {
			t.Errorf("endpointWithScheme(%q, %v) = %q, want %q", tc.endpoint, tc.disableSSL, got, tc.want)
		}
	}
}

// fakeS3 is an S3 endpoint recording the requests it receives
type fakeS3 struct {
	mu       sync.Mutex
	requests []fakeS3Request
}

type fakeS3Request struct {
	method    string
	path      string
	acl       bool
	cannedACL string
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, isACL := r.URL.Query()["acl"]

	f.mu.Lock()
	f.requests = append(f.requests, fakeS3Request{
		method:    r.Method,
		path:      r.URL.Path,
		acl:       isACL,
		cannedACL: r.Header.Get("x-amz-acl"),
	})
	f.mu.Unlock()

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", "3")
		w.Header().Set("Last-Modified", "Mon, 16 Nov 2026 10:00:00 GMT")
	}

	w.WriteHeader(http.StatusOK)
}

func (f *fakeS3) all() []fakeS3Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]fakeS3Request(nil), f.requests...)
}

func (f *fakeS3) aclRequests() []fakeS3Request {
	var out []fakeS3Request

	for _, req := range f.all() {
		if req.acl {
			out = append(out, req)
		}
	}

	return out
}

func loadFakeFs(t *testing.T, extraParams map[string]string) (afero.Fs, *fakeS3) {
	t.Helper()

	fake := &fakeS3{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	params := map[string]string{
		"endpoint":          server.URL,
		"region":            "us-east-1",
		"bucket":            "acme",
		"access_key_id":     "key",
		"secret_access_key": "secret",
		"path_style":        "true",
		"basePath":          "users/u",
	}
	for k, v := range extraParams {
		params[k] = v
	}

	fs, err := LoadFs(&confpar.Access{Fs: "s3", Params: params})
	if err != nil {
		t.Fatalf("LoadFs: %v", err)
	}

	return fs, fake
}

func TestChmodRefusedByDefault(t *testing.T) {
	fs, fake := loadFakeFs(t, nil)

	err := fs.Chmod("/report.pdf", 0o644)
	if !errors.Is(err, ErrChmodNotSupported) {
		t.Fatalf("Chmod error = %v, want %v", err, ErrChmodNotSupported)
	}

	if reqs := fake.aclRequests(); len(reqs) != 0 {
		t.Fatalf("ACL requests sent: %+v", reqs)
	}
}

func TestChmodACLOptIn(t *testing.T) {
	fs, fake := loadFakeFs(t, map[string]string{"chmod_acl": "true"})

	if err := fs.Chmod("/report.pdf", 0o644); err != nil {
		t.Fatalf("Chmod 644: %v", err)
	}

	if err := fs.Chmod("/report.pdf", 0o600); err != nil {
		t.Fatalf("Chmod 600: %v", err)
	}

	reqs := fake.aclRequests()
	if len(reqs) != 2 {
		t.Fatalf("got %d ACL requests, want 2: %+v", len(reqs), reqs)
	}

	for i, want := range []string{"public-read", "private"} {
		if reqs[i].method != http.MethodPut || reqs[i].path != "/acme/users/u/report.pdf" || reqs[i].cannedACL != want {
			t.Errorf("ACL request %d = %+v, want PUT /acme/users/u/report.pdf with %s", i, reqs[i], want)
		}
	}
}

func TestOtherOperationsReachBackend(t *testing.T) {
	fs, fake := loadFakeFs(t, nil)

	info, err := fs.Stat("/report.pdf")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if info.Size() != 3 {
		t.Errorf("Size = %d, want 3", info.Size())
	}

	if err := fs.Remove("/report.pdf"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	var sawHead, sawDelete bool

	for _, req := range fake.all() {
		if req.path != "/acme/users/u/report.pdf" {
			continue
		}

		switch req.method {
		case http.MethodHead:
			sawHead = true
		case http.MethodDelete:
			sawDelete = true
		}
	}

	if !sawHead || !sawDelete {
		t.Errorf("requests %+v: want a HEAD and a DELETE on /acme/users/u/report.pdf", fake.all())
	}
}
