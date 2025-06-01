package ping

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
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

// MockHTTPClient implements http.Client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestCheckArchivedDependenciesWithProgress(t *testing.T) {
	// Helper to create HTML with specific published date
	createHTML := func(dateStr string) string {
		return `<html><body><span data-test-id="UnitHeader-commitTime">` + dateStr + `</span></body></html>`
	}

	// Use a fixed date that's more than 2 years old to ensure it's always detected as archived
	oldDate := time.Now().AddDate(-3, 0, 0).Format("Jan 2, 2006")
	recentDate := time.Now().AddDate(0, -2, 0).Format("Jan 2, 2006")

	// Date that's 8 months old - will be archived with 6-month threshold but active with default 2-year threshold
	eightMonthsOldDate := time.Now().AddDate(0, -8, 0).Format("Jan 2, 2006")

	tests := []struct {
		name                 string
		dependencies         []parser.Dependency
		mockResponses        map[string]*http.Response
		expectedArchived     []string
		expectedActive       []string
		expectedErrors       []string
		progressCalls        int
		archiveReasonCheck   map[string]string
		unmaintainedDuration time.Duration // New field for custom duration tests
	}{
		{
			name: "Active dependencies",
			dependencies: []parser.Dependency{
				{Path: "github.com/active/repo1", Indirect: false},
				{Path: "github.com/active/repo2", Indirect: false},
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/active/repo1": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(recentDate))),
				},
				"https://pkg.go.dev/github.com/active/repo2": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(recentDate))),
				},
			},
			expectedArchived: []string{},
			expectedActive:   []string{"github.com/active/repo1", "github.com/active/repo2"},
			expectedErrors:   []string{},
			progressCalls:    2,
		},
		{
			name: "Archived dependencies (old)",
			dependencies: []parser.Dependency{
				{Path: "github.com/old/repo1", Indirect: false},
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/old/repo1": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(oldDate))),
				},
			},
			expectedArchived:   []string{"github.com/old/repo1"},
			expectedActive:     []string{},
			expectedErrors:     []string{},
			progressCalls:      1,
			archiveReasonCheck: map[string]string{"github.com/old/repo1": "Not updated since"},
		},
		{
			name: "Archived dependencies (404)",
			dependencies: []parser.Dependency{
				{Path: "github.com/notfound/repo", Indirect: false},
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/notfound/repo": {
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				},
			},
			expectedArchived:   []string{"github.com/notfound/repo"},
			expectedActive:     []string{},
			expectedErrors:     []string{},
			progressCalls:      1,
			archiveReasonCheck: map[string]string{"github.com/notfound/repo": "404 from pkg.go.dev"},
		},
		{
			name: "Mixed dependencies",
			dependencies: []parser.Dependency{
				{Path: "github.com/active/repo", Indirect: false},
				{Path: "github.com/old/repo", Indirect: false},
				{Path: "github.com/notfound/repo", Indirect: false},
				{Path: "github.com/indirect/repo", Indirect: true}, // Should be ignored
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/active/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(recentDate))),
				},
				"https://pkg.go.dev/github.com/old/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(oldDate))),
				},
				"https://pkg.go.dev/github.com/notfound/repo": {
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				},
			},
			expectedArchived: []string{"github.com/old/repo", "github.com/notfound/repo"},
			expectedActive:   []string{"github.com/active/repo"},
			expectedErrors:   []string{},
			progressCalls:    3, // Should be 3 because indirect deps are filtered out
		},
		{
			name: "Custom duration - 6 months threshold",
			dependencies: []parser.Dependency{
				{Path: "github.com/recent/repo", Indirect: false},   // 2 months old
				{Path: "github.com/moderate/repo", Indirect: false}, // 8 months old - should be archived
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/recent/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(recentDate))),
				},
				"https://pkg.go.dev/github.com/moderate/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(eightMonthsOldDate))),
				},
			},
			expectedArchived:     []string{"github.com/moderate/repo"},
			expectedActive:       []string{"github.com/recent/repo"},
			expectedErrors:       []string{},
			progressCalls:        2,
			unmaintainedDuration: 6 * 30 * 24 * time.Hour, // 6 months
			archiveReasonCheck:   map[string]string{"github.com/moderate/repo": "Not updated since"},
		},
		{
			name: "Custom duration - 1 year threshold",
			dependencies: []parser.Dependency{
				{Path: "github.com/recent/repo", Indirect: false},   // 2 months old
				{Path: "github.com/moderate/repo", Indirect: false}, // 8 months old - should be active with 1 year threshold
				{Path: "github.com/old/repo", Indirect: false},      // 3 years old - should be archived
			},
			mockResponses: map[string]*http.Response{
				"https://pkg.go.dev/github.com/recent/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(recentDate))),
				},
				"https://pkg.go.dev/github.com/moderate/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(eightMonthsOldDate))),
				},
				"https://pkg.go.dev/github.com/old/repo": {
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(createHTML(oldDate))),
				},
			},
			expectedArchived:     []string{"github.com/old/repo"},
			expectedActive:       []string{"github.com/recent/repo", "github.com/moderate/repo"},
			expectedErrors:       []string{},
			progressCalls:        3,
			unmaintainedDuration: 365 * 24 * time.Hour, // 1 year
			archiveReasonCheck:   map[string]string{"github.com/old/repo": "Not updated since"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a client with a mocked HTTP client
			client := NewClient()

			// Set custom unmaintained duration if specified
			if tc.unmaintainedDuration > 0 {
				client.SetUnmaintainedDuration(tc.unmaintainedDuration)
			}

			// Create a proper mock of http.Client with a mocked Transport
			mockTransport := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					url := req.URL.String()
					resp, ok := tc.mockResponses[url]
					if !ok {
						t.Fatalf("Unexpected request to %s", url)
					}
					return resp, nil
				},
			}
			client.httpClient.Transport = mockTransport

			// Track progress calls with mutex for thread safety
			var mu sync.Mutex
			progressCalls := 0
			progressMessages := make(map[string]string)

			// Set the progress callback using the new method
			client.SetProgressCallback(func(dependency string, status string) {
				mu.Lock()
				defer mu.Unlock()
				progressCalls++
				progressMessages[dependency] = status
			})

			// Run the function under test
			results := client.PingPackage(tc.dependencies)

			// Verify the number of progress calls
			if progressCalls != tc.progressCalls {
				t.Errorf("Expected %d progress calls, got %d", tc.progressCalls, progressCalls)
			}

			// Check archived dependencies
			archivedDeps := []string{}
			activeDeps := []string{}
			errorDeps := []string{}

			for _, result := range results {
				if result.Error != "" {
					errorDeps = append(errorDeps, result.ModulePath)
				} else if result.IsArchived {
					archivedDeps = append(archivedDeps, result.ModulePath)

					// Check reason if specified
					if tc.archiveReasonCheck != nil {
						expectedReason, ok := tc.archiveReasonCheck[result.ModulePath]
						if ok && !strings.Contains(result.Reason, expectedReason) {
							t.Errorf("For %s expected reason to contain %q, got %q",
								result.ModulePath, expectedReason, result.Reason)
						}
					}
				} else {
					activeDeps = append(activeDeps, result.ModulePath)
				}
			}

			// Helper function to check slices
			checkSlices := func(name string, got, expected []string) {
				if len(got) != len(expected) {
					t.Errorf("Expected %d %s, got %d", len(expected), name, len(got))
					t.Errorf("Expected: %v", expected)
					t.Errorf("Got: %v", got)
					return
				}

				// Convert to maps for easier comparison
				expectedMap := make(map[string]bool)
				for _, dep := range expected {
					expectedMap[dep] = true
				}

				for _, dep := range got {
					if !expectedMap[dep] {
						t.Errorf("Unexpected %s: %s", name, dep)
					}
				}
			}

			checkSlices("archived dependencies", archivedDeps, tc.expectedArchived)
			checkSlices("active dependencies", activeDeps, tc.expectedActive)
			checkSlices("error dependencies", errorDeps, tc.expectedErrors)
		})
	}
}
