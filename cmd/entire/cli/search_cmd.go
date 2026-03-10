package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/entireio/cli/cmd/entire/cli/jsonutil"
	"github.com/entireio/cli/cmd/entire/cli/search"
	"github.com/entireio/cli/cmd/entire/cli/strategy"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		jsonFlag   bool
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
OpenAI embeddings with BM25 full-text search.`,
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

			if !jsonFlag {
				fmt.Fprintf(cmd.ErrOrStderr(), "Searching %s/%s for: %s\n", owner, repoName, query)
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
				if jsonFlag {
					fmt.Fprintln(cmd.OutOrStdout(), "[]")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "No results found.")
				}
				return nil
			}

			if jsonFlag {
				data, err := jsonutil.MarshalIndentWithNewline(resp.Results, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling results: %w", err)
				}
				fmt.Fprint(cmd.OutOrStdout(), string(data))
				return nil
			}

			// Pretty print
			fmt.Fprintf(cmd.OutOrStdout(), "\nFound %d results for %s:\n\n", resp.Total, resp.Repo)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "RANK\tCHECKPOINT\tSCORE\tMATCH\tBRANCH\tAUTHOR\tPROMPT")
			for i, r := range resp.Results {
				branch := truncateStr(r.Branch, 20)
				author := "-"
				if r.CommitAuthorUsername != nil {
					author = *r.CommitAuthorUsername
				} else if r.CommitAuthor != nil {
					author = *r.CommitAuthor
				}
				prompt := "-"
				if r.Prompt != nil {
					prompt = truncateStr(*r.Prompt, 40)
				}
				fmt.Fprintf(w, "%d\t%s\t%.4f\t%s\t%s\t%s\t%s\n",
					i+1, truncateStr(r.CheckpointID, 12), r.SearchMeta.RRFScore,
					r.SearchMeta.MatchType, branch, author, prompt)
			}
			_ = w.Flush()
			fmt.Fprintln(cmd.OutOrStdout())

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.Flags().StringVar(&branchFlag, "branch", "", "Filter results by branch name")
	cmd.Flags().IntVar(&limitFlag, "limit", 20, "Maximum number of results")

	return cmd
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
