//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/entireio/cli/cmd/entire/cli/session"
)

// TestSubagentAccumulation_Issue591 reproduces the issue #591 shape: multiple
// subagent sessions end with uncommitted files that never overlap with a later
// parent-session commit. The stale ENDED subagent sessions must be force-condensed
// on that later commit so they do not accumulate and get re-processed forever.
func TestSubagentAccumulation_Issue591(t *testing.T) {
	t.Parallel()

	env := NewFeatureBranchEnv(t)

	type subagentInfo struct {
		SessionID string
		File      string
	}

	t.Log("Phase 1: start parent session and create stale ended subagent sessions")

	parent := env.NewSession()
	if err := env.SimulateUserPromptSubmit(parent.ID); err != nil {
		t.Fatalf("SimulateUserPromptSubmit for parent failed: %v", err)
	}
	parent.TranscriptBuilder.AddUserMessage("Use subagents to investigate the work and then create the parent file.")

	const numSubagents = 4
	subagents := make([]subagentInfo, 0, numSubagents)

	for i := 0; i < numSubagents; i++ {
		sub := env.NewSession()
		file := fmt.Sprintf("subagent_work_%d.go", i)
		content := fmt.Sprintf("package main\n\nfunc SubagentWork%d() {}\n", i)

		if err := env.SimulateUserPromptSubmit(sub.ID); err != nil {
			t.Fatalf("SimulateUserPromptSubmit for subagent %d failed: %v", i, err)
		}

		env.WriteFile(file, content)
		sub.CreateTranscript("Create "+file, []FileChange{
			{Path: file, Content: content},
		})

		if err := env.SimulateStop(sub.ID, sub.TranscriptPath); err != nil {
			t.Fatalf("SimulateStop for subagent %d failed: %v", i, err)
		}
		if err := env.SimulateSessionEnd(sub.ID); err != nil {
			t.Fatalf("SimulateSessionEnd for subagent %d failed: %v", i, err)
		}

		state, err := env.GetSessionState(sub.ID)
		if err != nil {
			t.Fatalf("GetSessionState for subagent %d failed: %v", i, err)
		}
		if state == nil {
			t.Fatalf("subagent %d session state missing", i)
		}
		if state.Phase != session.PhaseEnded {
			t.Fatalf("subagent %d phase = %s, want ended", i, state.Phase)
		}
		if len(state.FilesTouched) == 0 {
			t.Fatalf("subagent %d FilesTouched should be non-empty before cleanup", i)
		}
		if state.StepCount == 0 {
			t.Fatalf("subagent %d StepCount should be > 0 before cleanup", i)
		}
		if state.FullyCondensed {
			t.Fatalf("subagent %d should not be FullyCondensed before cleanup", i)
		}

		staleEndedAt := time.Now().Add(-2 * time.Hour)
		state.EndedAt = &staleEndedAt
		if err := env.WriteSessionState(sub.ID, state); err != nil {
			t.Fatalf("WriteSessionState for subagent %d failed: %v", i, err)
		}

		taskToolUseID := fmt.Sprintf("toolu_parent_subagent_%d", i)
		parent.TranscriptBuilder.AddTaskToolUse(taskToolUseID, "Create "+file)
		parent.TranscriptBuilder.AddTaskToolResult(taskToolUseID, sub.ID)

		subagents = append(subagents, subagentInfo{
			SessionID: sub.ID,
			File:      file,
		})
	}

	t.Log("Phase 2: parent commits unrelated work")

	parentFile := "parent_work.go"
	parentContent := "package main\n\nfunc ParentWork() {}\n"
	env.WriteFile(parentFile, parentContent)
	parentToolUseID := parent.TranscriptBuilder.AddToolUse("mcp__acp__Write", parentFile, parentContent)
	parent.TranscriptBuilder.AddToolResult(parentToolUseID)
	parent.TranscriptBuilder.AddAssistantMessage("Done.")
	if err := parent.TranscriptBuilder.WriteToFile(parent.TranscriptPath); err != nil {
		t.Fatalf("failed to write parent transcript: %v", err)
	}

	if err := env.SimulateStop(parent.ID, parent.TranscriptPath); err != nil {
		t.Fatalf("SimulateStop for parent failed: %v", err)
	}

	env.GitCommitWithShadowHooks("Parent commit", parentFile)

	t.Log("Phase 3: verify stale ended subagent sessions were force-condensed")

	for i, sub := range subagents {
		state, err := env.GetSessionState(sub.SessionID)
		if err != nil {
			t.Fatalf("GetSessionState after parent commit for subagent %d failed: %v", i, err)
		}
		if state == nil {
			t.Fatalf("subagent %d session state missing after parent commit", i)
		}
		if state.Phase != session.PhaseEnded {
			t.Fatalf("subagent %d phase after parent commit = %s, want ended", i, state.Phase)
		}
		if state.StepCount != 0 {
			t.Fatalf("subagent %d StepCount = %d, want 0 after force-condense", i, state.StepCount)
		}
		if len(state.FilesTouched) != 0 {
			t.Fatalf("subagent %d FilesTouched = %v, want empty after force-condense", i, state.FilesTouched)
		}
		if !state.FullyCondensed {
			t.Fatalf("subagent %d should be FullyCondensed after force-condense", i)
		}
	}

	t.Log("Phase 4: make another unrelated commit and verify fully-condensed subagents stay skipped")

	followUp := env.NewSession()
	if err := env.SimulateUserPromptSubmit(followUp.ID); err != nil {
		t.Fatalf("SimulateUserPromptSubmit for follow-up session failed: %v", err)
	}

	followUpFile := "follow_up.go"
	followUpContent := "package main\n\nfunc FollowUp() {}\n"
	env.WriteFile(followUpFile, followUpContent)
	followUp.CreateTranscript("Create "+followUpFile, []FileChange{
		{Path: followUpFile, Content: followUpContent},
	})
	if err := env.SimulateStop(followUp.ID, followUp.TranscriptPath); err != nil {
		t.Fatalf("SimulateStop for follow-up session failed: %v", err)
	}

	env.GitCommitWithShadowHooks("Follow-up commit", followUpFile)

	for i, sub := range subagents {
		state, err := env.GetSessionState(sub.SessionID)
		if err != nil {
			t.Fatalf("GetSessionState after follow-up commit for subagent %d failed: %v", i, err)
		}
		if state == nil {
			t.Fatalf("subagent %d session state missing after follow-up commit", i)
		}
		if !state.FullyCondensed {
			t.Fatalf("subagent %d should remain FullyCondensed after follow-up commit", i)
		}
		if state.StepCount != 0 {
			t.Fatalf("subagent %d StepCount = %d after follow-up commit, want 0", i, state.StepCount)
		}
	}
}
