package auth

import (
	"os"
	"path/filepath"
	"testing"
)

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

// chdirTo changes the process cwd and returns a cleanup function that restores it.
func chdirTo(t *testing.T, dir string) func() {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir(%q): %v", dir, err)
	}
	return func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restoring cwd to %q: %v", orig, err)
		}
	}
}

// TestAuthFilePathWalkUp verifies that authFilePath finds the .entire directory
// by walking up from a nested subdirectory.
func TestAuthFilePathWalkUp(t *testing.T) {
	root, subsubdir := setupTempRepoDir(t)
	defer chdirTo(t, subsubdir)()

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
		real, err := filepath.EvalSymlinks(filepath.Dir(p))
		if err != nil {
			t.Fatalf("EvalSymlinks(%q): %v", filepath.Dir(p), err)
		}
		return filepath.Join(real, filepath.Base(p))
	}
	if resolveDir(t, got) != resolveDir(t, want) {
		t.Errorf("authFilePath() = %q, want %q", got, want)
	}
}

// TestReadWriteAuthRoundTrip verifies that writeAuth followed by readAuth
// returns the same data.
func TestReadWriteAuthRoundTrip(t *testing.T) {
	_, subsubdir := setupTempRepoDir(t)
	defer chdirTo(t, subsubdir)()

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
	// Use a temp dir with a .entire directory but no auth.json inside.
	_, subsubdir := setupTempRepoDir(t)
	defer chdirTo(t, subsubdir)()

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
	_, subsubdir := setupTempRepoDir(t)
	defer chdirTo(t, subsubdir)()

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

// TestSetStoredTokenPreservesUsername verifies that setStoredToken does not
// overwrite an existing username.
func TestSetStoredTokenPreservesUsername(t *testing.T) {
	_, subsubdir := setupTempRepoDir(t)
	defer chdirTo(t, subsubdir)()

	// Pre-populate with a username.
	if err := setStoredUsername("carol"); err != nil {
		t.Fatalf("setStoredUsername: %v", err)
	}

	// Now set a token; username should be preserved.
	if err := setStoredToken("newtoken"); err != nil {
		t.Fatalf("setStoredToken: %v", err)
	}

	user, err := GetStoredUsername()
	if err != nil {
		t.Fatalf("GetStoredUsername: %v", err)
	}
	if user != "carol" {
		t.Errorf("username = %q, want %q", user, "carol")
	}

	tok, err := GetStoredToken()
	if err != nil {
		t.Fatalf("GetStoredToken: %v", err)
	}
	if tok != "newtoken" {
		t.Errorf("token = %q, want %q", tok, "newtoken")
	}
}
