package heartbeat

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
)

// RepoStatus contains information about a repository's status
type RepoStatus struct {
	ModulePath    string
	Owner         string
	Repo          string
	IsArchived    bool
	StatusCode    int
	Error         string
	LastPublished time.Time
	Reason        string
}

// Client is an HTTP client for checking module status
type Client struct {
	httpClient *http.Client
	token      string // Kept for backward compatibility
}

// NewClient creates a new client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

// ProgressCallback is a type for the progress callback function
type ProgressCallback func(dependency string, status string)

// CheckArchivedDependenciesWithProgress checks which dependencies appear to be archived
// by checking their status on pkg.go.dev
func (c *Client) CheckArchivedDependenciesWithProgress(
	deps []parser.Dependency,
	progress ProgressCallback,
) []RepoStatus {
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
				progress(dep.Path, "Error: "+err.Error())
			} else {
				// Primary check: Is the published date older than 2 year?
				if !publishDate.IsZero() && time.Since(publishDate) > 2*365*24*time.Hour {
					status.IsArchived = true
					status.Reason = fmt.Sprintf("Not updated since %s", publishDate.Format("Jan 2, 2006"))
					progress(dep.Path, "ARCHIVED (Last published: "+publishDate.Format("Jan 2, 2006")+")")
				} else if statusCode == http.StatusNotFound {
					// Secondary check: Is the package not found on pkg.go.dev?
					status.IsArchived = true
					status.Reason = "404 from pkg.go.dev"
					progress(dep.Path, "ARCHIVED (Not found on pkg.go.dev)")
				} else {
					// Recent publish date and status code is OK
					progress(dep.Path, "Active (Last published: "+publishDate.Format("Jan 2, 2006")+")")
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
	publishDate = extractPublishDate(htmlContent)

	return statusCode, repoURL, publishDate, nil
}

// extractPublishDate extracts the last published date from pkg.go.dev HTML
func extractPublishDate(html string) time.Time {
	// Look for the published date in the span with data-test-id="UnitHeader-commitTime"
	datePattern := regexp.MustCompile(`<span[^>]*data-test-id="UnitHeader-commitTime"[^>]*>([^<]+)</span>`)
	matches := datePattern.FindStringSubmatch(html)

	if len(matches) < 2 {
		return time.Time{} // Return zero time if not found
	}

	// Parse the date string (format: "Jan 23, 2024")
	dateStr := strings.TrimSpace(matches[1])
	dateStr = strings.TrimPrefix(dateStr, "Published:")
	dateStr = strings.TrimSpace(dateStr)

	// Try different formats as the exact format might vary
	formats := []string{
		"Jan 2, 2006",
		"Jan 02, 2006",
		"January 2, 2006",
		"January 02, 2006",
	}

	for _, format := range formats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date
		}
	}

	// If parsing fails with standard formats, try a more flexible approach
	// Extract month, day, year
	parts := strings.Split(dateStr, " ")
	if len(parts) >= 3 {
		// Try to build a standardized date string
		month := parts[0]
		day := strings.TrimSuffix(parts[1], ",")
		year := parts[2]

		// Ensure day is 2 digits
		if len(day) == 1 {
			day = "0" + day
		}

		// Try parsing again
		standardized := fmt.Sprintf("%s %s, %s", month, day, year)
		date, err := time.Parse("Jan 02, 2006", standardized)
		if err == nil {
			return date
		}
	}

	return time.Time{} // Return zero time if parsing fails
}
