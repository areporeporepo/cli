package auth

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestMain forces the file backend for all tests. The keyring backend requires
// an OS keyring daemon which is not available in CI or test environments.
func TestMain(m *testing.M) {
	os.Setenv("ENTIRE_TOKEN_STORE", "file") //nolint:tenv // set before any test runs; intentional global
	os.Exit(m.Run())
}

// resetBackend resets the backend singleton so each test starts with a clean slate.
func resetBackend() {
	once = sync.Once{}
	backend = nil
}

// setupTempRepoDir creates a temp dir with a .entire subdirectory and returns:
//   - the root of the temp dir (where .entire lives)
//   - a deeply-nested subdirectory to simulate a project subdirectory
func setupTempRepoDir(t *testing.T) (root, subsubdir string) {
	t.Helper()
	root = t.TempDir()
	entireDirPath := filepath.Join(root, entireDir)
	if err := os.MkdirAll(entireDirPath, 0o700); err != nil {
		t.Fatalf("creating .entire dir: %v", err)
	}
	subsubdir = filepath.Join(root, "subdir", "subsubdir")
	if err := os.MkdirAll(subsubdir, 0o755); err != nil {
		t.Fatalf("creating subsubdir: %v", err)
	}
	return root, subsubdir
}

// chdirTo changes the process cwd for the duration of the test.
// Restoration is handled automatically by t.Chdir.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	t.Chdir(dir)
}

// TestAuthFilePathWalkUp verifies that authFilePath finds the .entire directory
// by walking up from a nested subdirectory.
func TestAuthFilePathWalkUp(t *testing.T) {
	root, subsubdir := setupTempRepoDir(t)
	chdirTo(t, subsubdir)

	got, err := authFilePath()
	if err != nil {
		t.Fatalf("authFilePath() error: %v", err)
	}

	want := filepath.Join(root, entireDir, authFileName)

	// Resolve symlinks on the directory portions only (the file itself doesn't exist yet).
	// On macOS /var is a symlink to /private/var, so os.Getwd() and t.TempDir() may
	// return different-looking but equivalent paths.
	resolveDir := func(t *testing.T, p string) string {
		t.Helper()
		resolved, err := filepath.EvalSymlinks(filepath.Dir(p))
		if err != nil {
			t.Fatalf("EvalSymlinks(%q): %v", filepath.Dir(p), err)
		}
		return filepath.Join(resolved, filepath.Base(p))
	}
	if resolveDir(t, got) != resolveDir(t, want) {
		t.Errorf("authFilePath() = %q, want %q", got, want)
	}
}

// TestReadWriteAuthRoundTrip verifies that writeAuth followed by readAuth
// returns the same data.
func TestReadWriteAuthRoundTrip(t *testing.T) {
	_, subsubdir := setupTempRepoDir(t)
	chdirTo(t, subsubdir)

	in := &storedAuth{Token: "tok123", Username: "alice"}
	if err := writeAuth(in); err != nil {
		t.Fatalf("writeAuth: %v", err)
	}

	out, err := readAuth()
	if err != nil {
		t.Fatalf("readAuth: %v", err)
	}
	if out.Token != in.Token {
		t.Errorf("Token = %q, want %q", out.Token, in.Token)
	}
	if out.Username != in.Username {
		t.Errorf("Username = %q, want %q", out.Username, in.Username)
	}
}

// TestGetStoredTokenNoFile verifies that GetStoredToken returns ("", nil) when
// no auth file exists.
func TestGetStoredTokenNoFile(t *testing.T) {
	resetBackend()
	_, subsubdir := setupTempRepoDir(t)
	chdirTo(t, subsubdir)

	tok, err := GetStoredToken()
	if err != nil {
		t.Fatalf("GetStoredToken() unexpected error: %v", err)
	}
	if tok != "" {
		t.Errorf("GetStoredToken() = %q, want empty string", tok)
	}
}

// TestSetStoredAuthWritesBothFields verifies that SetStoredAuth stores both
// token and username in a single operation.
func TestSetStoredAuthWritesBothFields(t *testing.T) {
	resetBackend()
	_, subsubdir := setupTempRepoDir(t)
	chdirTo(t, subsubdir)

	if err := SetStoredAuth("mytoken", "bob"); err != nil {
		t.Fatalf("SetStoredAuth: %v", err)
	}

	tok, err := GetStoredToken()
	if err != nil {
		t.Fatalf("GetStoredToken: %v", err)
	}
	if tok != "mytoken" {
		t.Errorf("token = %q, want %q", tok, "mytoken")
	}

	user, err := GetStoredUsername()
	if err != nil {
		t.Fatalf("GetStoredUsername: %v", err)
	}
	if user != "bob" {
		t.Errorf("username = %q, want %q", user, "bob")
	}
}

// TestSetStoredAuthPreservesExistingUsername verifies that calling SetStoredAuth
// with a new token still stores the username supplied in the same call.
func TestSetStoredAuthPreservesExistingUsername(t *testing.T) {
	resetBackend()
	_, subsubdir := setupTempRepoDir(t)
	chdirTo(t, subsubdir)

	// First login: both token and username.
	if err := SetStoredAuth("token-v1", "carol"); err != nil {
		t.Fatalf("SetStoredAuth (first): %v", err)
	}

	// Re-auth with a new token for the same user.
	if err := SetStoredAuth("token-v2", "carol"); err != nil {
		t.Fatalf("SetStoredAuth (second): %v", err)
	}

	tok, err := GetStoredToken()
	if err != nil {
		t.Fatalf("GetStoredToken: %v", err)
	}
	if tok != "token-v2" {
		t.Errorf("token = %q, want %q", tok, "token-v2")
	}

	user, err := GetStoredUsername()
	if err != nil {
		t.Fatalf("GetStoredUsername: %v", err)
	}
	if user != "carol" {
		t.Errorf("username = %q, want %q", user, "carol")
	}
}

// TestTokenSourceFileBackend verifies TokenSource returns the file backend name
// when ENTIRE_TOKEN_STORE=file is set.
func TestTokenSourceFileBackend(t *testing.T) {
	resetBackend()
	if got := TokenSource(); got != SourceEntireDir {
		t.Errorf("TokenSource() = %q, want %q", got, SourceEntireDir)
	}
}
