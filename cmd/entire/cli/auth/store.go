// Package auth implements credential storage for the Entire CLI.
//
// By default tokens are stored in the OS keyring (macOS Keychain, Linux Secret
// Service, Windows Credential Manager). Set ENTIRE_TOKEN_STORE=file to use a
// JSON file instead, which is useful in CI environments that lack a keyring daemon.
//
// When using the file backend tokens are stored in .entire/auth.json, discovered
// by walking up from the current working directory.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	// SourceEntireDir is the display name for the file-based token store.
	SourceEntireDir = ".entire/auth.json"
	// SourceKeyring is the display name for the OS keyring token store.
	SourceKeyring = "OS keyring"

	entireDir    = ".entire"
	authFileName = "auth.json"

	keyringService      = "entire-cli"
	keyringTokenKey     = "github-token"
	keyringUsernameKey  = "github-username"
)

// errNoAuth is returned by the file store when no auth file exists.
var errNoAuth = errors.New("no auth file")

var (
	once    sync.Once
	backend tokenStore
)

// tokenStore is the interface for pluggable credential storage.
type tokenStore interface {
	GetToken() (string, error)
	GetUsername() (string, error)
	SetAuth(token, username string) error
	DeleteAuth() error
	Source() string
}

func resolveBackend() {
	once.Do(func() {
		if os.Getenv("ENTIRE_TOKEN_STORE") == "file" {
			backend = fileTokenStore{}
		} else {
			backend = keyringTokenStore{}
		}
	})
}

// GetStoredToken retrieves the GitHub token. Returns ("", nil) if not stored.
func GetStoredToken() (string, error) {
	resolveBackend()
	return backend.GetToken()
}

// GetStoredUsername retrieves the stored GitHub username. Returns ("", nil) if not stored.
func GetStoredUsername() (string, error) {
	resolveBackend()
	return backend.GetUsername()
}

// SetStoredAuth stores both the GitHub token and username atomically.
func SetStoredAuth(token, username string) error {
	resolveBackend()
	return backend.SetAuth(token, username)
}

// DeleteStoredToken removes all stored credentials.
func DeleteStoredToken() error {
	resolveBackend()
	return backend.DeleteAuth()
}

// TokenSource returns the display name of the active credential store.
func TokenSource() string {
	resolveBackend()
	return backend.Source()
}

// ─── Keyring backend ──────────────────────────────────────────────────────────

type keyringTokenStore struct{}

func (keyringTokenStore) GetToken() (string, error) {
	tok, err := keyring.Get(keyringService, keyringTokenKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading token from keyring: %w", err)
	}
	return tok, nil
}

func (keyringTokenStore) GetUsername() (string, error) {
	u, err := keyring.Get(keyringService, keyringUsernameKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading username from keyring: %w", err)
	}
	return u, nil
}

func (keyringTokenStore) SetAuth(token, username string) error {
	if err := keyring.Set(keyringService, keyringTokenKey, token); err != nil {
		return fmt.Errorf("storing token in keyring: %w", err)
	}
	if username != "" {
		if err := keyring.Set(keyringService, keyringUsernameKey, username); err != nil {
			return fmt.Errorf("storing username in keyring: %w", err)
		}
	}
	return nil
}

func (keyringTokenStore) DeleteAuth() error {
	if err := keyring.Delete(keyringService, keyringTokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("deleting token from keyring: %w", err)
	}
	if err := keyring.Delete(keyringService, keyringUsernameKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("deleting username from keyring: %w", err)
	}
	return nil
}

func (keyringTokenStore) Source() string { return SourceKeyring }

// ─── File backend ─────────────────────────────────────────────────────────────

type fileTokenStore struct{}

type storedAuth struct {
	Token    string `json:"token"`
	Username string `json:"username,omitempty"`
}

// authFilePath returns the path to .entire/auth.json by walking up from cwd.
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

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
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

func (fileTokenStore) GetToken() (string, error) {
	a, err := readAuth()
	if errors.Is(err, errNoAuth) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return a.Token, nil
}

func (fileTokenStore) GetUsername() (string, error) {
	a, err := readAuth()
	if errors.Is(err, errNoAuth) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return a.Username, nil
}

func (fileTokenStore) SetAuth(token, username string) error {
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

func (fileTokenStore) DeleteAuth() error {
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

func (fileTokenStore) Source() string { return SourceEntireDir }
