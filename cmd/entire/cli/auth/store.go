package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	entireDir    = ".entire"
	authFileName = "auth.json"
)

// errNoAuth is returned by readAuth when no auth file exists.
var errNoAuth = errors.New("no auth file")

type storedAuth struct {
	Token    string `json:"token"`
	Username string `json:"username,omitempty"`
}

// authFilePath returns the path to .entire/auth.json in the current repo root.
// Walks up from cwd to find the .entire directory.
func authFilePath() (string, error) {
	dir, err := os.Getwd() //nolint:forbidigo // walks up to find .entire, handles subdirectory case explicitly
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, entireDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return filepath.Join(candidate, authFileName), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fall back to cwd/.entire/auth.json
	return filepath.Join(entireDir, authFileName), nil
}

func readAuth() (*storedAuth, error) {
	path, err := authFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) //nolint:gosec // reading from controlled .entire path
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errNoAuth
	}
	if err != nil {
		return nil, fmt.Errorf("reading auth file: %w", err)
	}

	var a storedAuth
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parsing auth file: %w", err)
	}

	return &a, nil
}

func writeAuth(a *storedAuth) error {
	path, err := authFilePath()
	if err != nil {
		return err
	}

	// Ensure .entire directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling auth: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing auth file: %w", err)
	}

	return nil
}

// GetStoredToken retrieves the GitHub token from .entire/auth.json.
// Returns ("", nil) if no token is stored.
func GetStoredToken() (string, error) {
	a, err := readAuth()
	if errors.Is(err, errNoAuth) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return a.Token, nil
}

// SetStoredToken stores the GitHub token in .entire/auth.json.
func SetStoredToken(token string) error {
	a, err := readAuth()
	if err != nil && !errors.Is(err, errNoAuth) {
		return fmt.Errorf("reading existing auth: %w", err)
	}
	if a == nil {
		a = &storedAuth{}
	}
	a.Token = token
	return writeAuth(a)
}

// SetStoredAuth stores both the GitHub token and username atomically in a single write.
func SetStoredAuth(token, username string) error {
	a, err := readAuth()
	if errors.Is(err, errNoAuth) || a == nil {
		a = &storedAuth{}
	} else if err != nil {
		return fmt.Errorf("reading existing auth: %w", err)
	}
	a.Token = token
	a.Username = username
	return writeAuth(a)
}

// DeleteStoredToken removes the auth file.
func DeleteStoredToken() error {
	path, err := authFilePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err //nolint:wrapcheck // os error is descriptive enough
}

// GetStoredUsername retrieves the stored GitHub username.
// Returns ("", nil) if no username is stored.
func GetStoredUsername() (string, error) {
	a, err := readAuth()
	if errors.Is(err, errNoAuth) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return a.Username, nil
}

// SetStoredUsername stores the GitHub username in .entire/auth.json.
func SetStoredUsername(username string) error {
	a, err := readAuth()
	if err != nil && !errors.Is(err, errNoAuth) {
		return fmt.Errorf("reading existing auth: %w", err)
	}
	if a == nil {
		a = &storedAuth{}
	}
	a.Username = username
	return writeAuth(a)
}
