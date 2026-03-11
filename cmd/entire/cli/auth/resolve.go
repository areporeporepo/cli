package auth

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// Source indicates where a resolved token came from.
type Source string

const (
	SourceEntireDir   Source = ".entire/auth.json"
	SourceEnvironment Source = "GITHUB_TOKEN"
	SourceGHCLI       Source = "gh auth token"
)

// ResolveGitHubToken resolves a GitHub token from available sources.
// Resolution order: .entire/auth.json → GITHUB_TOKEN env → gh auth token.
func ResolveGitHubToken(ctx context.Context) (token string, source Source, err error) {
	// 1. .entire/auth.json (entire login)
	token, err = GetStoredToken()
	if err == nil && token != "" {
		return token, SourceEntireDir, nil
	}

	// 2. GITHUB_TOKEN environment variable
	token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token != "" {
		return token, SourceEnvironment, nil
	}

	// 3. gh CLI
	out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
	if err == nil {
		token = strings.TrimSpace(string(out))
		if token != "" {
			return token, SourceGHCLI, nil
		}
	}

	return "", "", nil
}
