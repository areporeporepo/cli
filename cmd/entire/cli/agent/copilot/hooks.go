package copilot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/entireio/cli/cmd/entire/cli/agent"
	"github.com/entireio/cli/cmd/entire/cli/paths"
)

// Ensure CopilotAgent implements HookSupport and HookHandler
var (
	_ agent.HookSupport = (*CopilotAgent)(nil)
	_ agent.HookHandler = (*CopilotAgent)(nil)
)

// Copilot hook names - these become subcommands under `entire hooks github-copilot`
const (
	HookNameSessionStart        = "session-start"
	HookNameSessionEnd          = "session-end"
	HookNameUserPromptSubmitted = "user-prompt-submitted"
	HookNameAgentStop           = "agent-stop"
	HookNameSubagentStop        = "subagent-stop"
	HookNamePreToolUse          = "pre-tool-use"
	HookNamePostToolUse         = "post-tool-use"
	HookNameErrorOccurred       = "error-occurred"
)

// CopilotConfigFileName is the hook config file used by GitHub Copilot.
const CopilotConfigFileName = "copilot-setup.json"

// hookNameToConfigKey maps CLI subcommand names to Copilot's native hook names (camelCase in config)
var hookNameToConfigKey = map[string]string{
	HookNameSessionStart:        "sessionStart",
	HookNameSessionEnd:          "sessionEnd",
	HookNameUserPromptSubmitted: "userPromptSubmitted",
	HookNameAgentStop:           "agentStop",
	HookNameSubagentStop:        "subagentStop",
	HookNamePreToolUse:          "preToolUse",
	HookNamePostToolUse:         "postToolUse",
	HookNameErrorOccurred:       "errorOccurred",
}

// entireHookPrefixes are command prefixes that identify Entire hooks
var entireHookPrefixes = []string{
	"entire ",
	"go run ${COPILOT_PROJECT_DIR}/cmd/entire/main.go ",
}

// GetHookNames implements agent.HookHandler by delegating to HookNames.
func (c *CopilotAgent) GetHookNames() []string {
	return c.HookNames()
}

// InstallHooks installs Copilot hooks in .github/hooks/copilot-setup.json.
// If force is true, removes existing Entire hooks before installing.
// Returns the number of hooks installed.
func (c *CopilotAgent) InstallHooks(localDev bool, force bool) (int, error) {
	repoRoot, err := paths.RepoRoot()
	if err != nil {
		repoRoot, err = os.Getwd() //nolint:forbidigo // Intentional fallback when RepoRoot() fails (tests run outside git repos)
		if err != nil {
			return 0, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	configPath := filepath.Join(repoRoot, ".github", "hooks", CopilotConfigFileName)

	// Read existing config if it exists
	var config hookConfig
	existingData, readErr := os.ReadFile(configPath) //nolint:gosec // path is constructed from repo root + fixed path
	if readErr == nil {
		if err := json.Unmarshal(existingData, &config); err != nil {
			return 0, fmt.Errorf("failed to parse existing %s: %w", CopilotConfigFileName, err)
		}
	} else {
		config.Version = 1
	}

	if config.Hooks == nil {
		config.Hooks = make(map[string][]hookEntry)
	}

	// Define hook command based on localDev mode
	var cmdPrefix string
	if localDev {
		cmdPrefix = "go run ${COPILOT_PROJECT_DIR}/cmd/entire/main.go hooks github-copilot "
	} else {
		cmdPrefix = "entire hooks github-copilot "
	}

	// Check for idempotency BEFORE removing hooks
	if !force {
		configKey := hookNameToConfigKey[HookNameSessionStart]
		expectedCmd := cmdPrefix + HookNameSessionStart
		if entries, ok := config.Hooks[configKey]; ok {
			for _, entry := range entries {
				if entry.Bash == expectedCmd {
					return 0, nil // Already installed with same mode
				}
			}
		}
	}

	// Remove existing Entire hooks first
	for configKey, entries := range config.Hooks {
		config.Hooks[configKey] = removeEntireHookEntries(entries)
		if len(config.Hooks[configKey]) == 0 {
			delete(config.Hooks, configKey)
		}
	}

	// Install all 8 hooks
	hookNames := c.HookNames()
	for _, hookName := range hookNames {
		configKey, ok := hookNameToConfigKey[hookName]
		if !ok {
			continue
		}

		entry := hookEntry{
			Type:       "command",
			Bash:       cmdPrefix + hookName,
			TimeoutSec: 10,
		}

		config.Hooks[configKey] = append(config.Hooks[configKey], entry)
	}

	count := len(hookNames)

	// Write config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return 0, fmt.Errorf("failed to create .github/hooks directory: %w", err)
	}

	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, output, 0o600); err != nil {
		return 0, fmt.Errorf("failed to write %s: %w", CopilotConfigFileName, err)
	}

	return count, nil
}

// UninstallHooks removes Entire hooks from Copilot config.
func (c *CopilotAgent) UninstallHooks() error {
	repoRoot, err := paths.RepoRoot()
	if err != nil {
		repoRoot = "."
	}

	configPath := filepath.Join(repoRoot, ".github", "hooks", CopilotConfigFileName)
	data, err := os.ReadFile(configPath) //nolint:gosec // path is constructed from repo root + fixed path
	if err != nil {
		return nil //nolint:nilerr // No config file means nothing to uninstall
	}

	var config hookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse %s: %w", CopilotConfigFileName, err)
	}

	if config.Hooks == nil {
		return nil
	}

	// Remove Entire hooks from all hook types
	for configKey, entries := range config.Hooks {
		config.Hooks[configKey] = removeEntireHookEntries(entries)
		if len(config.Hooks[configKey]) == 0 {
			delete(config.Hooks, configKey)
		}
	}

	// Write back
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, output, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", CopilotConfigFileName, err)
	}

	return nil
}

// AreHooksInstalled checks if Entire hooks are installed in Copilot config.
func (c *CopilotAgent) AreHooksInstalled() bool {
	repoRoot, err := paths.RepoRoot()
	if err != nil {
		repoRoot = "."
	}

	configPath := filepath.Join(repoRoot, ".github", "hooks", CopilotConfigFileName)
	data, err := os.ReadFile(configPath) //nolint:gosec // path is constructed from repo root + fixed path
	if err != nil {
		return false
	}

	var config hookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	for _, entries := range config.Hooks {
		for _, entry := range entries {
			if isEntireHook(entry.Bash) {
				return true
			}
		}
	}

	return false
}

// GetSupportedHooks returns the hook types Copilot supports.
func (c *CopilotAgent) GetSupportedHooks() []agent.HookType {
	return []agent.HookType{
		agent.HookSessionStart,
		agent.HookSessionEnd,
		agent.HookStop,
		agent.HookUserPromptSubmit,
		agent.HookPreToolUse,
		agent.HookPostToolUse,
	}
}

// Helper functions

// isEntireHook checks if a command is an Entire hook
func isEntireHook(command string) bool {
	for _, prefix := range entireHookPrefixes {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	return false
}

// removeEntireHookEntries removes all Entire hook entries from a slice
func removeEntireHookEntries(entries []hookEntry) []hookEntry {
	result := make([]hookEntry, 0, len(entries))
	for _, entry := range entries {
		if !isEntireHook(entry.Bash) {
			result = append(result, entry)
		}
	}
	return result
}
