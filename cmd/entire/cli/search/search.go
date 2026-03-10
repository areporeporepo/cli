package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiTimeout = 30 * time.Second

// DefaultServiceURL is the production search service URL.
const DefaultServiceURL = "https://entire.io"

// Result represents a single search result from the search service.
type Result struct {
	CheckpointID  string  `json:"checkpoint_id"`
	RRF           float64 `json:"rrf"`
	VectorRank    *int    `json:"vectorRank"`
	BM25Rank      *int    `json:"bm25Rank"`
	MatchType     string  `json:"matchType"`
	Branch        *string `json:"branch"`
	Agent         *string `json:"agent"`
	Author        *string `json:"author"`
	CreatedAt     *string `json:"created_at"`
	CommitSHA     *string `json:"commit_sha"`
	CommitMessage *string `json:"commit_message"`
	Prompt        *string `json:"prompt"`
	FilesTouched  *string `json:"files_touched"`
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
		q.Set("limit", fmt.Sprintf("%d", cfg.Limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+cfg.GitHubToken)
	req.Header.Set("User-Agent", "entire-cli")

	client := &http.Client{}
	resp, err := client.Do(req) //nolint:bodyclose,gosec // closed below; URL is constructed from trusted config
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
