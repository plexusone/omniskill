// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package installer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// GitHubFetcher fetches skill packages from GitHub releases.
type GitHubFetcher struct {
	httpClient *http.Client
	token      string
}

// NewGitHubFetcher creates a new GitHub fetcher.
func NewGitHubFetcher() *GitHubFetcher {
	return &GitHubFetcher{
		httpClient: &http.Client{},
		token:      os.Getenv("GITHUB_TOKEN"),
	}
}

// WithToken sets the GitHub token for authenticated requests.
func (f *GitHubFetcher) WithToken(token string) *GitHubFetcher {
	f.token = token
	return f
}

// GetLatestRelease gets the latest release for a repository.
func (f *GitHubFetcher) GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return f.getRelease(ctx, url)
}

// GetRelease gets a specific release by tag.
func (f *GitHubFetcher) GetRelease(ctx context.Context, owner, repo, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	return f.getRelease(ctx, url)
}

// getRelease fetches a release from the given URL.
func (f *GitHubFetcher) getRelease(ctx context.Context, url string) (*GitHubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &release, nil
}

// DownloadAndExtract downloads a release tarball and extracts it to the target directory.
func (f *GitHubFetcher) DownloadAndExtract(ctx context.Context, release *GitHubRelease, targetDir string) error {
	// Create the target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	// Download tarball
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, release.TarballURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed (status %d)", resp.StatusCode)
	}

	// Extract tarball
	return extractTarGz(resp.Body, targetDir)
}

// extractTarGz extracts a gzipped tarball to the target directory.
func extractTarGz(r io.Reader, targetDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	var rootDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		// GitHub tarballs have a root directory like "owner-repo-hash"
		// We need to strip this when extracting
		parts := strings.SplitN(header.Name, "/", 2)
		if rootDir == "" && len(parts) > 0 {
			rootDir = parts[0]
		}

		// Skip the root directory entry
		if header.Name == rootDir || header.Name == rootDir+"/" {
			continue
		}

		// Calculate the relative path without the root directory
		var relPath string
		if len(parts) > 1 {
			relPath = parts[1]
		} else {
			continue
		}

		targetPath := filepath.Join(targetDir, relPath)

		// Extract only the permission bits (lower 12 bits) to avoid int64->uint32 overflow
		mode := os.FileMode(header.Mode & 0o7777)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, mode); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory: %w", err)
			}

			// Create the file
			//nolint:gosec // G110: Decompression bomb protection not needed for GitHub releases
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}

			//nolint:gosec // G110: Decompression bomb protection not needed for GitHub releases
			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("write file: %w", err)
			}
			_ = outFile.Close()

			// Set file permissions
			if err := os.Chmod(targetPath, mode); err != nil {
				return fmt.Errorf("set permissions: %w", err)
			}
		}
	}

	return nil
}

// ParseGitHubRef parses a GitHub repository reference.
// Formats:
//   - github.com/owner/repo
//   - github.com/owner/repo@tag
//   - owner/repo
//   - owner/repo@tag
func ParseGitHubRef(ref string) (owner, repo, tag string, err error) {
	// Strip github.com prefix if present
	ref = strings.TrimPrefix(ref, "github.com/")

	// Split off tag if present
	parts := strings.SplitN(ref, "@", 2)
	if len(parts) == 2 {
		tag = parts[1]
	}
	ref = parts[0]

	// Split owner and repo
	parts = strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid GitHub reference: %s", ref)
	}

	owner = parts[0]
	repo = parts[1]

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	return owner, repo, tag, nil
}
