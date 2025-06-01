package ping

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
	"github.com/Bhupesh-V/godeping/parsers/pkginfo"
)

// RepoStatus contains information about a repository's status
type RepoStatus struct {
	ModulePath    string    `json:"module_path"`
	Owner         string    `json:"-"`
	Repo          string    `json:"-"`
	IsArchived    bool      `json:"-"`
	StatusCode    int       `json:"-"`
	Error         string    `json:"-"`
	LastPublished time.Time `json:"last_published"`
	Reason        string    `json:"-"`
}

// Client is an HTTP client for checking module status
type Client struct {
	httpClient           *http.Client
	unmaintainedDuration time.Duration // Duration after which a module is considered unmaintained
	progress             func(dependency string, status string)
}

// NewClient creates a new client
func NewClient() *Client {
	return &Client{
		httpClient:           &http.Client{Timeout: 10 * time.Minute},
		unmaintainedDuration: 2 * 365 * 24 * time.Hour, // Default: 2 years
	}
}

// SetUnmaintainedDuration sets the duration threshold for considering a dependency unmaintained
func (c *Client) SetUnmaintainedDuration(d time.Duration) {
	c.unmaintainedDuration = d
}

// SetProgressCallback sets the callback function to report progress
func (c *Client) SetProgressCallback(callback func(dependency string, status string)) {
	c.progress = callback
}

// PingPackage checks which dependencies appear to be archived by checking their status on pkg.go.dev
func (c *Client) PingPackage(deps []parser.Dependency) []RepoStatus {
	// Filter out indirect dependencies
	var directDeps []parser.Dependency
	for _, dep := range deps {
		if !dep.Indirect {
			directDeps = append(directDeps, dep)
		}
	}

	resultChan := make(chan RepoStatus, len(directDeps))
	var wg sync.WaitGroup

	// Create a semaphore channel to limit concurrent requests to 10
	semaphore := make(chan struct{}, 10)

	// Launch a goroutine for each dependency
	for _, dep := range directDeps {
		wg.Add(1)
		go func(dep parser.Dependency) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Initialize status
			status := RepoStatus{
				ModulePath: dep.Path,
			}

			// Check package status on pkg.go.dev
			statusCode, _, publishDate, err := c.checkPackageStatus(dep.Path)
			status.StatusCode = statusCode
			status.LastPublished = publishDate

			if err != nil {
				status.Error = err.Error()
				c.progress(dep.Path, "Error: "+err.Error())
			} else {
				// Primary check: Is the published date older than the configured duration?
				if !publishDate.IsZero() && time.Since(publishDate) > c.unmaintainedDuration {
					status.IsArchived = true
					status.Reason = fmt.Sprintf("Not updated since %s", publishDate.Format("Jan 2, 2006"))
					c.progress(dep.Path, "Archived (Last published: "+publishDate.Format("Jan 2, 2006")+")")
				} else if statusCode == http.StatusNotFound {
					// Secondary check: Is the package not found on pkg.go.dev?
					status.IsArchived = true
					status.Reason = "404 from pkg.go.dev"
					c.progress(dep.Path, "Archived (Not found on pkg.go.dev)")
				} else {
					// Recent publish date and status code is OK
					c.progress(dep.Path, "Active (Last published: "+publishDate.Format("Jan 2, 2006")+")")
				}
			}

			resultChan <- status
		}(dep)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]RepoStatus, 0, len(directDeps))
	for status := range resultChan {
		results = append(results, status)
	}

	return results
}

// checkPackageStatus checks if a package exists on pkg.go.dev and extracts info
func (c *Client) checkPackageStatus(pkgPath string) (statusCode int, repoURL string, publishDate time.Time, err error) {
	url := fmt.Sprintf("https://pkg.go.dev/%s", pkgPath)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, "", time.Time{}, err
	}
	defer resp.Body.Close()

	// Return the status code regardless
	statusCode = resp.StatusCode

	// If it's not a successful response, return early
	if resp.StatusCode != http.StatusOK {
		return statusCode, "", time.Time{}, nil
	}

	// Try to extract repository URL and publish date from HTML
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statusCode, "", time.Time{}, err
	}

	htmlContent := string(body)
	publishDate = pkginfo.ExtractPublishDate(htmlContent)

	return statusCode, repoURL, publishDate, nil
}
