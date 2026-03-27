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
	MatchType string  `json:"matchType"`
	Score     float64 `json:"score"`
	Snippet   string  `json:"snippet,omitempty"`
}

// CheckpointResult represents a checkpoint returned by the search service.
type CheckpointResult struct {
	ID             string   `json:"id"`
	Prompt         string   `json:"prompt"`
	CommitMessage  *string  `json:"commitMessage"`
	CommitSHA      *string  `json:"commitSha"`
	Branch         string   `json:"branch"`
	Org            string   `json:"org"`
	Repo           string   `json:"repo"`
	Author         string   `json:"author"`
	AuthorUsername *string  `json:"authorUsername"`
	CreatedAt      string   `json:"createdAt"`
	FilesTouched   []string `json:"filesTouched"`
}

// Result wraps a search result with its type and ranking metadata.
type Result struct {
	Type string           `json:"type"`
	Data CheckpointResult `json:"data"`
	Meta Meta             `json:"searchMeta"`
}

// Response is the search service response.
type Response struct {
	Results []Result `json:"results"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Error   string   `json:"error,omitempty"`
}

// Config holds the configuration for a search request.
type Config struct {
	ServiceURL  string // Base URL of the search service
	GitHubToken string
	Owner       string
	Repo        string
	Query       string
	Limit       int
}

var httpClient = &http.Client{Timeout: apiTimeout}

// Search calls the search service to perform a hybrid search.
func Search(ctx context.Context, cfg Config) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	serviceURL := cfg.ServiceURL
	if serviceURL == "" {
		serviceURL = DefaultServiceURL
	}

	u, err := url.Parse(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("parsing service URL: %w", err)
	}
	u.Path = "/search/v1/search"

	q := u.Query()
	q.Set("q", cfg.Query)
	q.Set("repo", cfg.Owner+"/"+cfg.Repo)
	q.Set("types", "checkpoints")
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

	resp, err := httpClient.Do(req)
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

	if result.Error != "" {
		return nil, fmt.Errorf("search service error: %s", result.Error)
	}

	return &result, nil
}
