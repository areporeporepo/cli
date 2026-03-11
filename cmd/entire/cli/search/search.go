package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const apiTimeout = 30 * time.Second

// DefaultServiceURL is the production search service URL.
const DefaultServiceURL = "https://entire.io"

// Meta contains search ranking metadata for a result.
type Meta struct {
	RRFScore   float64 `json:"rrfScore"`
	MatchType  string  `json:"matchType"`
	VectorRank *int    `json:"vectorRank"`
	BM25Rank   *int    `json:"bm25Rank"`
}

// Result represents a single search result from the search service.
type Result struct {
	CheckpointID         string      `json:"checkpointId"`
	Branch               string      `json:"branch"`
	CommitSHA            *string     `json:"commitSha"`
	CommitMessage        *string     `json:"commitMessage"`
	CommitAuthor         *string     `json:"commitAuthor"`
	CommitAuthorUsername *string     `json:"commitAuthorUsername"`
	CommitDate           *string     `json:"commitDate"`
	Additions            int         `json:"additions"`
	Deletions            int         `json:"deletions"`
	FilesChanged         int         `json:"filesChanged"`
	FilesTouched         []string    `json:"filesTouched"`
	FileStats            interface{} `json:"fileStats"`
	Prompt               *string     `json:"prompt"`
	Agent                string      `json:"agent"`
	Steps                int         `json:"steps"`
	SessionCount         int         `json:"sessionCount"`
	CreatedAt            string      `json:"createdAt"`
	InputTokens          *int        `json:"inputTokens"`
	OutputTokens         *int        `json:"outputTokens"`
	CacheCreationTokens  *int        `json:"cacheCreationTokens"`
	CacheReadTokens      *int        `json:"cacheReadTokens"`
	APICallCount         *int        `json:"apiCallCount"`
	Meta                 Meta        `json:"searchMeta"`
}

// Response is the search service response.
type Response struct {
	Results []Result `json:"results"`
	Query   string   `json:"query"`
	Repo    string   `json:"repo"`
	Total   int      `json:"total"`
	Error   string   `json:"error,omitempty"`
}

// Config holds the configuration for a search request.
type Config struct {
	ServiceURL  string // Base URL of the search service
	GitHubToken string
	Owner       string
	Repo        string
	Query       string
	Branch      string
	Limit       int
}

// Search calls the search service to perform a hybrid search.
func Search(ctx context.Context, cfg Config) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	serviceURL := cfg.ServiceURL
	if serviceURL == "" {
		serviceURL = DefaultServiceURL
	}

	// Build URL: /search/v1/:owner/:repo?q=...&branch=...&limit=...
	u, err := url.Parse(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("parsing service URL: %w", err)
	}
	u.Path = fmt.Sprintf("/search/v1/%s/%s", url.PathEscape(cfg.Owner), url.PathEscape(cfg.Repo))

	q := u.Query()
	q.Set("q", cfg.Query)
	if cfg.Branch != "" {
		q.Set("branch", cfg.Branch)
	}
	if cfg.Limit > 0 {
		q.Set("limit", strconv.Itoa(cfg.Limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+cfg.GitHubToken)
	req.Header.Set("User-Agent", "entire-cli")

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is constructed from trusted config
	if err != nil {
		return nil, fmt.Errorf("calling search service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("search service error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("search service returned %d: %s", resp.StatusCode, string(body))
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}
