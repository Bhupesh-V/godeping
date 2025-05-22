package ping

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 10*time.Minute, client.httpClient.Timeout)
}

func TestCheckPackageStatus(t *testing.T) {
	tests := []struct {
		name           string
		pkgPath        string
		serverResponse string
		statusCode     int
		expectError    bool
		publishDate    time.Time
	}{
		{
			name:           "Successful response",
			pkgPath:        "github.com/example/pkg",
			serverResponse: `<span data-test-id="UnitHeader-commitTime">Jan 15, 2023</span>`,
			statusCode:     http.StatusOK,
			expectError:    false,
			publishDate:    time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:           "Not found response",
			pkgPath:        "github.com/notexist/pkg",
			serverResponse: "Not Found",
			statusCode:     http.StatusNotFound,
			expectError:    false,
			publishDate:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create a client with custom transport that redirects to our test server
			client := NewClient()
			originalTransport := client.httpClient.Transport
			client.httpClient.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = server.Listener.Addr().String()
				if originalTransport == nil {
					return http.DefaultTransport.RoundTrip(req)
				}
				return originalTransport.RoundTrip(req)
			})

			// Call the function
			statusCode, _, publishDate, err := client.checkPackageStatus(tt.pkgPath)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.statusCode, statusCode)
				assert.Equal(t, tt.publishDate, publishDate)
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
