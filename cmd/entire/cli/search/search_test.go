package search

import (
	"testing"
)

func TestParseGitHubRemote_SSH(t *testing.T) {
	t.Parallel()
	owner, repo, err := ParseGitHubRemote("git@github.com:entirehq/entire.io.git")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "entirehq" || repo != "entire.io" {
		t.Errorf("got %s/%s, want entirehq/entire.io", owner, repo)
	}
}

func TestParseGitHubRemote_HTTPS(t *testing.T) {
	t.Parallel()
	owner, repo, err := ParseGitHubRemote("https://github.com/entirehq/entire.io.git")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "entirehq" || repo != "entire.io" {
		t.Errorf("got %s/%s, want entirehq/entire.io", owner, repo)
	}
}

func TestParseGitHubRemote_HTTPSNoGit(t *testing.T) {
	t.Parallel()
	owner, repo, err := ParseGitHubRemote("https://github.com/entirehq/entire.io")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "entirehq" || repo != "entire.io" {
		t.Errorf("got %s/%s, want entirehq/entire.io", owner, repo)
	}
}

func TestParseGitHubRemote_Invalid(t *testing.T) {
	t.Parallel()
	_, _, err := ParseGitHubRemote("")
	if err == nil {
		t.Error("expected error for empty URL")
	}

	_, _, err = ParseGitHubRemote("not-a-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
