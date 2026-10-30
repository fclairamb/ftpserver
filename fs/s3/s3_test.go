package s3

import "testing"

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
