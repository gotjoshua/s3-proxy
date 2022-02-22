//go:build unit

package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
)

func TestGetRequestURI(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:989/fake/path", nil)
	if err != nil {
		t.Fatal(err)
	}

	want := "http://localhost:989/fake/path"
	got := GetRequestURI(req)
	if got != want {
		t.Errorf("GetRequestURI() = %v, want %v", got, want)
	}
}

func TestProxiedGetRequestURI(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:989/fake/path", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add the same header a Load Balancer should set
	req.Header.Set("X-Forwarded-Proto", "https")

	want := "https://localhost:989/fake/path"
	got := GetRequestURI(req)
	if got != want {
		t.Errorf("GetRequestURI() = %v, want %v", got, want)
	}
}

func Test_GetRequestHost(t *testing.T) {
	hXForwardedHost1 := http.Header{
		"X-Forwarded-Host": []string{"fake.host"},
	}
	hXForwardedHost2 := http.Header{
		"X-Forwarded-Host": []string{"fake.host:9090"},
	}
	hForwarded := http.Header{
		"Forwarded": []string{"for=192.0.2.60;proto=http;by=203.0.113.43;host=fake.host:9090"},
	}

	tests := []struct {
		name     string
		headers  http.Header
		inputURL string
		want     string
	}{
		{
			name:     "request host",
			headers:  nil,
			inputURL: "http://request.host/",
			want:     "request.host",
		},
		{
			name:     "forwarded host",
			headers:  hForwarded,
			inputURL: "http://fake.host:9090/",
			want:     "fake.host:9090",
		},
		{
			name:     "x-forwarded host 1",
			headers:  hXForwardedHost1,
			inputURL: "http://fake.host/",
			want:     "fake.host",
		},
		{
			name:     "x-forwarded host 2",
			headers:  hXForwardedHost2,
			inputURL: "http://fake.host:9090/",
			want:     "fake.host:9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.inputURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.headers != nil {
				req.Header = tt.headers
			}

			if got := GetRequestHost(req); got != tt.want {
				t.Errorf("RequestHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequestScheme(t *testing.T) {
	hForwardedHttps := http.Header{
		"Forwarded": []string{"for=192.0.2.60;proto=https;by=203.0.113.43;host=fake.host:9090"},
	}
	hForwardedHttp := http.Header{
		"Forwarded": []string{"for=192.0.2.60;proto=http;by=203.0.113.43;host=fake.host:9090"},
	}
	hXForwardedProtoHttps := http.Header{
		"X-Forwarded-Proto": []string{"https"},
	}
	hXForwardedProtoHttp := http.Header{
		"X-Forwarded-Proto": []string{"http"},
	}
	tests := []struct {
		name    string
		headers http.Header
		want    string
	}{
		{
			name:    "Forwarded HTTPS",
			headers: hForwardedHttps,
			want:    "https",
		},
		{
			name:    "Forwarded HTTP",
			headers: hForwardedHttp,
			want:    "http",
		},
		{
			name:    "X-Forwarded-Proto HTTPS",
			headers: hXForwardedProtoHttps,
			want:    "https",
		},
		{
			name:    "X-Forwarded-Proto HTTP",
			headers: hXForwardedProtoHttp,
			want:    "http",
		},
		{
			name: "None",
			want: "http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://fake.host:9090/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.headers != nil {
				req.Header = tt.headers
			}
			if got := GetRequestScheme(req); got != tt.want {
				t.Errorf("GetRequestScheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTLSVersion(t *testing.T) {
	var tlsString string
	var tlsVersion uint16

	for _, prefix := range []string{"1", "TLS1", "TLSv1", "TLS-1", "TLS_1", "TLS 1", "tls1", "tlsv1", "tls-1", "tls_1"} {
		for _, separator := range []string{".", "-", "_"} {
			tlsString = fmt.Sprintf("%s%s0", prefix, separator)
			tlsVersion = ParseTLSVersion(tlsString)

			if tlsVersion != tls.VersionTLS10 {
				t.Errorf("ParseTLSVersion(%#v) = %v, want %v", tlsString, tlsVersion, tls.VersionTLS10)
			}

			tlsString = fmt.Sprintf("%s%s1", prefix, separator)
			tlsVersion = ParseTLSVersion(tlsString)

			if tlsVersion != tls.VersionTLS11 {
				t.Errorf("ParseTLSVersion(%#v) = %v, want %v", tlsString, tlsVersion, tls.VersionTLS11)
			}

			tlsString = fmt.Sprintf("%s%s2", prefix, separator)
			tlsVersion = ParseTLSVersion(tlsString)

			if tlsVersion != tls.VersionTLS12 {
				t.Errorf("ParseTLSVersion(%#v) = %v, want %v", tlsString, tlsVersion, tls.VersionTLS12)
			}

			tlsString = fmt.Sprintf("%s%s3", prefix, separator)
			tlsVersion = ParseTLSVersion(tlsString)

			if tlsVersion != tls.VersionTLS13 {
				t.Errorf("ParseTLSVersion(%#v) = %v, want %v", tlsString, tlsVersion, tls.VersionTLS13)
			}
		}
	}

	if ParseTLSVersion("") != 0 {
		t.Errorf("Expected ParseTLSVersion(\"\") to return 0")
	}

	if ParseTLSVersion("TLS") != 0 {
		t.Errorf("Expected ParseTLSVersion(\"TLS\") to return 0")
	}

	if ParseTLSVersion("TLS&1.1") != 0 {
		t.Errorf("Expected ParseTLSVersion(\"TLS&1.1\") to return 0")
	}

	if ParseTLSVersion("TLS-1+1") != 0 {
		t.Errorf("Expected ParseTLSVersion(\"TLS-1+1\") to return 0")
	}

	if ParseTLSVersion("TLSv1.9") != 0 {
		t.Errorf("Expected ParseTLSVersion(\"TLSv1.9\") to return 0")
	}
}
