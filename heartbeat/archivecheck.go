package githubchecker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Bhupesh-V/godepbeat/parser"
)

// RepoStatus contains information about a repository's status
type RepoStatus struct {
	ModulePath string
	Owner      string
	Repo       string
	IsArchived bool
	Error      string
}

// GitHubResponse represents the relevant fields from GitHub API response
type GitHubResponse struct {
	Archived bool `json:"archived"`
}

// Client is a GitHub API client with rate limiting
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}
}

// ProgressCallback is a type for the progress callback function
type ProgressCallback func(dependency string, status string)

// CheckArchivedDependenciesWithProgress checks which GitHub-hosted dependencies are archived
// and reports final status via callback
func (c *Client) CheckArchivedDependenciesWithProgress(
	deps []parser.Dependency,
	progress ProgressCallback,
) []RepoStatus {
	results := make([]RepoStatus, 0, len(deps))

	// Filter out indirect dependencies
	var directDeps []parser.Dependency
	for _, dep := range deps {
		if !dep.Indirect {
			directDeps = append(directDeps, dep)
		}
	}

	// Use a simple rate limiter to avoid hitting GitHub API limits
	rateLimiter := time.Tick(time.Second / 10) // Max 10 requests per second

	for _, dep := range directDeps {
		// Wait for rate limiter
		<-rateLimiter

		// Check if it's a GitHub dependency
		owner, repo, isGitHub := parseGitHubPath(dep.Path)
		if !isGitHub {
			progress(dep.Path, "Not a GitHub dependency")
			continue
		}

		status := RepoStatus{
			ModulePath: dep.Path,
			Owner:      owner,
			Repo:       repo,
		}

		// Check if archived
		isArchived, err := c.isRepoArchived(owner, repo)
		if err != nil {
			status.Error = err.Error()
			progress(dep.Path, "Error: "+err.Error())
		} else {
			status.IsArchived = isArchived
			if isArchived {
				progress(dep.Path, "ARCHIVED")
			} else {
				progress(dep.Path, "Active")
			}
		}

		results = append(results, status)
	}

	return results
}

// isRepoArchived checks if a GitHub repository is archived
func (c *Client) isRepoArchived(owner, repo string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Add authentication if token is provided
	if c.token != "" {
		req.Header.Add("Authorization", "token "+c.token)
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("User-Agent", "godepbeat")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var ghResp GitHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return false, err
	}

	return ghResp.Archived, nil
}

// parseGitHubPath extracts owner and repo from a GitHub module path
func parseGitHubPath(path string) (owner, repo string, isGitHub bool) {
	// Check if it's a GitHub URL
	if !strings.HasPrefix(path, "github.com/") {
		return "", "", false
	}

	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "", "", false
	}

	// Handle submodules and packages
	owner = parts[1]
	repo = parts[2]

	// Handle repos with .git suffix
	repo = strings.TrimSuffix(repo, ".git")

	// Remove version suffix (e.g., /v2, /v3) from repo name
	repo = removeVersionSuffix(repo)

	return owner, repo, true
}

// removeVersionSuffix removes version suffixes like /v2, /v3 from repo names
func removeVersionSuffix(repo string) string {
	// Match common version patterns like v1, v2, etc.
	versionPattern := regexp.MustCompile(`^(.*?)(/v\d+)?$`)
	matches := versionPattern.FindStringSubmatch(repo)

	if len(matches) > 1 {
		return matches[1] // Return the repo name without version suffix

	}

	return repo
}
