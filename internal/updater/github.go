package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubAPIBase = "https://api.github.com"
	repoOwner     = "cstone-io"
	repoName      = "twine"
	userAgent     = "twine-cli-updater"
)

// GitHubRelease represents a GitHub release with the fields we care about.
type GitHubRelease struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []GitHubAsset  `json:"assets"`
}

// GitHubAsset represents a downloadable asset from a release.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// GitHubClient handles communication with the GitHub API.
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: githubAPIBase,
	}
}

// GetLatestRelease fetches the latest non-prerelease version from GitHub.
func (c *GitHubClient) GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, repoOwner, repoName)

	release := &GitHubRelease{}
	if err := c.doRequest(url, release); err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	return release, nil
}

// GetRelease fetches a specific release by tag name.
func (c *GitHubClient) GetRelease(tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.baseURL, repoOwner, repoName, tag)

	release := &GitHubRelease{}
	if err := c.doRequest(url, release); err != nil {
		return nil, fmt.Errorf("failed to fetch release %s: %w", tag, err)
	}

	return release, nil
}

// ListReleases fetches all releases (up to 100, which should be plenty).
func (c *GitHubClient) ListReleases() ([]GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", c.baseURL, repoOwner, repoName)

	var releases []GitHubRelease
	if err := c.doRequest(url, &releases); err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	return releases, nil
}

// DownloadAsset downloads a binary asset from a URL and returns the data.
func (c *GitHubClient) DownloadAsset(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download data: %w", err)
	}

	return data, nil
}

// doRequest performs a JSON API request and decodes the response.
func (c *GitHubClient) doRequest(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource not found (404)")
	}

	if resp.StatusCode == http.StatusForbidden {
		// Check if it's rate limiting
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return fmt.Errorf("GitHub API rate limit exceeded")
		}
		return fmt.Errorf("access forbidden (403)")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
