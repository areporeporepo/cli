package copilot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHooks_FreshInstall(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}
	count, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("InstallHooks() error = %v", err)
	}

	if count != 8 {
		t.Errorf("InstallHooks() count = %d, want 8", count)
	}

	config := readCopilotConfig(t, tempDir)

	if config.Version != 1 {
		t.Errorf("version = %d, want 1", config.Version)
	}

	// Verify all 8 hook types are present
	expectedConfigKeys := []string{
		"sessionStart", "sessionEnd", "userPromptSubmitted", "agentStop",
		"subagentStop", "preToolUse", "postToolUse", "errorOccurred",
	}
	for _, key := range expectedConfigKeys {
		entries, ok := config.Hooks[key]
		if !ok {
			t.Errorf("hook %q not found in config", key)
			continue
		}
		if len(entries) != 1 {
			t.Errorf("hook %q has %d entries, want 1", key, len(entries))
		}
	}

	// Verify command format
	if entries, ok := config.Hooks["sessionStart"]; ok && len(entries) > 0 {
		expected := "entire hooks github-copilot session-start"
		if entries[0].Bash != expected {
			t.Errorf("sessionStart bash = %q, want %q", entries[0].Bash, expected)
		}
		if entries[0].Type != "command" {
			t.Errorf("sessionStart type = %q, want command", entries[0].Type)
		}
		if entries[0].TimeoutSec != 10 {
			t.Errorf("sessionStart timeoutSec = %d, want 10", entries[0].TimeoutSec)
		}
	}
}

func TestInstallHooks_LocalDev(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}
	_, err := ag.InstallHooks(true, false)
	if err != nil {
		t.Fatalf("InstallHooks() error = %v", err)
	}

	config := readCopilotConfig(t, tempDir)

	// Verify local dev commands use go run
	if entries, ok := config.Hooks["sessionStart"]; ok && len(entries) > 0 {
		expected := "go run ${COPILOT_PROJECT_DIR}/cmd/entire/main.go hooks github-copilot session-start"
		if entries[0].Bash != expected {
			t.Errorf("sessionStart bash = %q, want %q", entries[0].Bash, expected)
		}
	}
}

func TestInstallHooks_Idempotent(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}

	count1, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("first InstallHooks() error = %v", err)
	}
	if count1 != 8 {
		t.Errorf("first InstallHooks() count = %d, want 8", count1)
	}

	count2, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("second InstallHooks() error = %v", err)
	}
	if count2 != 0 {
		t.Errorf("second InstallHooks() count = %d, want 0 (idempotent)", count2)
	}

	// Verify still only 1 entry per hook type
	config := readCopilotConfig(t, tempDir)
	for key, entries := range config.Hooks {
		if len(entries) != 1 {
			t.Errorf("hook %q has %d entries after double install, want 1", key, len(entries))
		}
	}
}

func TestInstallHooks_Force(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}

	_, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("first InstallHooks() error = %v", err)
	}

	count, err := ag.InstallHooks(false, true)
	if err != nil {
		t.Fatalf("force InstallHooks() error = %v", err)
	}
	if count != 8 {
		t.Errorf("force InstallHooks() count = %d, want 8", count)
	}
}

func TestInstallHooks_PreservesUserHooks(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create config with existing user hooks
	writeCopilotConfig(t, tempDir, `{
  "version": 1,
  "hooks": {
    "sessionStart": [
      {"type": "command", "bash": "echo user-hook", "timeoutSec": 5}
    ]
  }
}`)

	ag := &CopilotAgent{}
	_, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("InstallHooks() error = %v", err)
	}

	config := readCopilotConfig(t, tempDir)

	entries, ok := config.Hooks["sessionStart"]
	if !ok {
		t.Fatal("sessionStart not found")
	}
	if len(entries) != 2 {
		t.Errorf("sessionStart entries = %d, want 2 (user + entire)", len(entries))
	}

	// Verify user hook is preserved
	foundUserHook := false
	for _, entry := range entries {
		if entry.Bash == "echo user-hook" {
			foundUserHook = true
		}
	}
	if !foundUserHook {
		t.Error("user hook was not preserved")
	}
}

func TestUninstallHooks(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}

	_, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("InstallHooks() error = %v", err)
	}

	if !ag.AreHooksInstalled() {
		t.Error("hooks should be installed before uninstall")
	}

	if err := ag.UninstallHooks(); err != nil {
		t.Fatalf("UninstallHooks() error = %v", err)
	}

	if ag.AreHooksInstalled() {
		t.Error("hooks should not be installed after uninstall")
	}
}

func TestUninstallHooks_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}

	err := ag.UninstallHooks()
	if err != nil {
		t.Fatalf("UninstallHooks() should not error when no config file: %v", err)
	}
}

func TestUninstallHooks_PreservesUserHooks(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create config with both user and entire hooks
	writeCopilotConfig(t, tempDir, `{
  "version": 1,
  "hooks": {
    "sessionStart": [
      {"type": "command", "bash": "echo user-hook", "timeoutSec": 5},
      {"type": "command", "bash": "entire hooks github-copilot session-start", "timeoutSec": 10}
    ]
  }
}`)

	ag := &CopilotAgent{}
	if err := ag.UninstallHooks(); err != nil {
		t.Fatalf("UninstallHooks() error = %v", err)
	}

	config := readCopilotConfig(t, tempDir)

	entries, ok := config.Hooks["sessionStart"]
	if !ok {
		t.Fatal("sessionStart should still exist with user hook")
	}
	if len(entries) != 1 {
		t.Errorf("sessionStart entries = %d, want 1 (user only)", len(entries))
	}
	if entries[0].Bash != "echo user-hook" {
		t.Errorf("remaining hook = %q, want echo user-hook", entries[0].Bash)
	}
}

func TestAreHooksInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}

	if ag.AreHooksInstalled() {
		t.Error("AreHooksInstalled() should be false when no config file")
	}

	_, err := ag.InstallHooks(false, false)
	if err != nil {
		t.Fatalf("InstallHooks() error = %v", err)
	}

	if !ag.AreHooksInstalled() {
		t.Error("AreHooksInstalled() should be true after installation")
	}
}

func TestHookNames(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	names := ag.HookNames()

	expected := []string{
		HookNameSessionStart,
		HookNameSessionEnd,
		HookNameUserPromptSubmitted,
		HookNameAgentStop,
		HookNameSubagentStop,
		HookNamePreToolUse,
		HookNamePostToolUse,
		HookNameErrorOccurred,
	}

	if len(names) != len(expected) {
		t.Errorf("HookNames() returned %d names, want %d", len(names), len(expected))
	}

	for i, name := range expected {
		if names[i] != name {
			t.Errorf("HookNames()[%d] = %q, want %q", i, names[i], name)
		}
	}
}

// Helper functions

func readCopilotConfig(t *testing.T, tempDir string) hookConfig {
	t.Helper()
	configPath := filepath.Join(tempDir, ".github", "hooks", CopilotConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var config hookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	return config
}

func writeCopilotConfig(t *testing.T, tempDir, content string) {
	t.Helper()
	hooksDir := filepath.Join(tempDir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("failed to create hooks dir: %v", err)
	}
	configPath := filepath.Join(hooksDir, CopilotConfigFileName)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}
