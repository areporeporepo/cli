package strategy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/entireio/cli/cmd/entire/cli/paths"
	"github.com/entireio/cli/cmd/entire/cli/session"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostCommit_Issue591_SubagentScaleRegression is a regression test for GitHub
// issue #591: accumulated stale ENDED sessions caused O(N) overhead on every commit
// (~73-103ms per session, indefinitely). After the fix, they are force-condensed on
// the first PostCommit and skipped via FullyCondensed on all subsequent commits.
//
// Behavioral coverage (FullyCondensed/StepCount/FilesTouched assertions) lives in
// TestPostCommit_EndedSessionCarryForward_ForceCondensedWithoutOverlap. This test
// focuses on the performance contract: the second PostCommit must be significantly
// faster than the first once sessions are marked FullyCondensed.
func TestPostCommit_Issue591_SubagentScaleRegression(t *testing.T) {
	const sessionCount = 10

	dir := setupGitRepo(t)
	t.Chdir(dir)
	paths.ClearWorktreeRootCache()

	repo, err := git.PlainOpen(dir)
	require.NoError(t, err)

	s := &ManualCommitStrategy{}

	// Create sessions with shadow branches, then mark them stale ENDED.
	// Each session's FilesTouched = ["test.txt"], which won't overlap with the
	// unrelated.txt commit below, triggering the force-condense path.
	for i := range sessionCount {
		sessionID := fmt.Sprintf("ended-session-%d", i)
		setupSessionWithCheckpoint(t, s, repo, dir, sessionID)
		state, err := s.loadSessionState(context.Background(), sessionID)
		require.NoError(t, err)
		staleAt := time.Now().Add(-(forceCondenseThreshold + time.Minute))
		state.Phase = session.PhaseEnded
		state.EndedAt = &staleAt
		require.NoError(t, s.saveSessionState(context.Background(), state))
	}

	// Commit an unrelated file — no overlap with any session's FilesTouched.
	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "unrelated.txt"), []byte("user work"), 0o644))
	_, err = wt.Add("unrelated.txt")
	require.NoError(t, err)
	_, err = wt.Commit("unrelated\n\nEntire-Checkpoint: a1b2c3d4e5f6\n", &git.CommitOptions{
		Author: &object.Signature{Name: "User", Email: "user@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	paths.ClearWorktreeRootCache()

	// First PostCommit: force-condenses all stale ENDED sessions.
	firstStart := time.Now()
	require.NoError(t, s.PostCommit(context.Background()))
	firstElapsed := time.Since(firstStart)

	// Create second commit — all sessions are now FullyCondensed and should be skipped.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "second.txt"), []byte("second"), 0o644))
	_, err = wt.Add("second.txt")
	require.NoError(t, err)
	_, err = wt.Commit("second\n\nEntire-Checkpoint: b1b2b3b4b5b6\n", &git.CommitOptions{
		Author: &object.Signature{Name: "User", Email: "user@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	paths.ClearWorktreeRootCache()

	s2 := &ManualCommitStrategy{}
	secondStart := time.Now()
	require.NoError(t, s2.PostCommit(context.Background()))
	secondElapsed := time.Since(secondStart)

	// Second PostCommit must be significantly faster — sessions are FullyCondensed and skipped.
	assert.Less(t, secondElapsed, firstElapsed/2,
		"second PostCommit should be much faster once sessions are FullyCondensed (issue #591 regression)")
	t.Logf("first PostCommit (%d stale sessions): %v, second (all skipped): %v",
		sessionCount, firstElapsed, secondElapsed)
}
