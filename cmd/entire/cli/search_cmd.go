package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/entireio/cli/cmd/entire/cli/jsonutil"
	"github.com/entireio/cli/cmd/entire/cli/search"
	"github.com/entireio/cli/cmd/entire/cli/strategy"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		branchFlag string
		limitFlag  int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search checkpoints using semantic and keyword matching",
		Long: `Search checkpoints using hybrid search (semantic + keyword),
powered by the Entire search service.

Requires a GitHub token for authentication. The token is resolved from:
  1. GITHUB_TOKEN environment variable
  2. gh auth token (GitHub CLI)

Results are ranked using Reciprocal Rank Fusion (RRF) combining
OpenAI embeddings with BM25 full-text search.

Output is JSON by default for easy consumption by agents and scripts.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			query := strings.Join(args, " ")

			// Resolve GitHub token
			ghToken := os.Getenv("GITHUB_TOKEN")
			if ghToken == "" {
				// Try gh CLI
				out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
				if err == nil {
					ghToken = strings.TrimSpace(string(out))
				}
			}
			if ghToken == "" {
				return fmt.Errorf("GitHub token required. Set GITHUB_TOKEN or install gh CLI (gh auth login)")
			}

			// Get the repo's GitHub remote URL
			repo, err := strategy.OpenRepository(ctx)
			if err != nil {
				cmd.SilenceUsage = true
				fmt.Fprintln(cmd.ErrOrStderr(), "Not a git repository. Run this command from within a git repository.")
				return NewSilentError(err)
			}

			remote, err := repo.Remote("origin")
			if err != nil {
				return fmt.Errorf("could not find 'origin' remote: %w", err)
			}
			urls := remote.Config().URLs
			if len(urls) == 0 {
				return fmt.Errorf("origin remote has no URLs configured")
			}

			owner, repoName, err := search.ParseGitHubRemote(urls[0])
			if err != nil {
				return fmt.Errorf("parsing remote URL: %w", err)
			}

			serviceURL := os.Getenv("ENTIRE_SEARCH_URL")
			if serviceURL == "" {
				serviceURL = search.DefaultServiceURL
			}

			resp, err := search.Search(ctx, search.Config{
				ServiceURL:  serviceURL,
				GitHubToken: ghToken,
				Owner:       owner,
				Repo:        repoName,
				Query:       query,
				Branch:      branchFlag,
				Limit:       limitFlag,
			})
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(resp.Results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "[]")
				return nil
			}

			data, err := jsonutil.MarshalIndentWithNewline(resp.Results, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling results: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&branchFlag, "branch", "", "Filter results by branch name")
	cmd.Flags().IntVar(&limitFlag, "limit", 20, "Maximum number of results")

	return cmd
}
