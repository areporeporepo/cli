// Package search provides search functionality via the Entire search service.
package search

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ParseGitHubRemote extracts owner and repo from a GitHub remote URL.
// Supports SSH (git@github.com:owner/repo.git) and HTTPS (https://github.com/owner/repo.git).
func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return "", "", errors.New("empty remote URL")
	}

	var path string

	// SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		idx := strings.Index(remoteURL, ":")
		if idx < 0 {
			return "", "", fmt.Errorf("invalid SSH remote URL: %s", remoteURL)
		}
		host := remoteURL[len("git@"):idx]
		if host != "github.com" {
			return "", "", fmt.Errorf("remote is not a GitHub repository (host: %s)", host)
		}
		path = remoteURL[idx+1:]
	} else {
		// HTTPS format: https://github.com/owner/repo.git
		u, parseErr := url.Parse(remoteURL)
		if parseErr != nil {
			return "", "", fmt.Errorf("parsing remote URL: %w", parseErr)
		}
		if u.Host != "github.com" {
			return "", "", fmt.Errorf("remote is not a GitHub repository (host: %s)", u.Host)
		}
		path = strings.TrimPrefix(u.Path, "/")
	}

	// Remove .git suffix
	path = strings.TrimSuffix(path, ".git")

	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("could not extract owner/repo from remote URL: %s", remoteURL)
	}

	return parts[0], parts[1], nil
}
