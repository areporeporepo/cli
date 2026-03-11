// Package auth implements GitHub OAuth device flow authentication for the CLI.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	githubDeviceCodeURL = "https://github.com/login/device/code"
	githubTokenURL      = "https://github.com/login/oauth/access_token" //nolint:gosec // G101: OAuth endpoint URL, not a credential
	githubUserURL       = "https://api.github.com/user"

	// Scopes: read:user for identity, repo for private repo access checks.
	defaultScopes = "read:user repo"
)

// DeviceCodeResponse is the response from GitHub's device code endpoint.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int64  `json:"interval"`
}

// TokenResponse is the response from GitHub's token endpoint.
type TokenResponse struct {
	AccessToken string `json:"access_token"` //nolint:gosec // G117: field holds an OAuth token, pattern match is a false positive
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

var (
	// ErrAuthorizationPending means the user hasn't authorized yet.
	ErrAuthorizationPending = errors.New("authorization_pending")
	// ErrDeviceCodeExpired means the device code has expired.
	ErrDeviceCodeExpired = errors.New("device code expired")
	// ErrSlowDown means the client is polling too frequently.
	ErrSlowDown = errors.New("slow_down")
	// ErrAccessDenied means the user denied the authorization request.
	ErrAccessDenied = errors.New("access_denied")
)

// RequestDeviceCode initiates the GitHub device authorization flow.
func RequestDeviceCode(ctx context.Context, clientID string) (*DeviceCodeResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", defaultScopes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubDeviceCodeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is a constant
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed (%d): %s", resp.StatusCode, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &deviceResp, nil
}

// pollForToken polls GitHub's token endpoint once.
func pollForToken(ctx context.Context, clientID, deviceCode string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is a constant
	if err != nil {
		return nil, fmt.Errorf("polling for token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// GitHub returns 200 for both success and pending states
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if errStr, ok := raw["error"].(string); ok {
		switch errStr {
		case "authorization_pending":
			return nil, ErrAuthorizationPending
		case "slow_down":
			return nil, ErrSlowDown
		case "expired_token":
			return nil, ErrDeviceCodeExpired
		case "access_denied":
			return nil, ErrAccessDenied
		default:
			var desc string
			if d, ok := raw["error_description"].(string); ok {
				desc = d
			}
			return nil, fmt.Errorf("token request failed: %s - %s", errStr, desc)
		}
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, errors.New("empty access token in response")
	}

	return &tokenResp, nil
}

// WaitForAuthorization polls until the user authorizes or the code expires.
func WaitForAuthorization(ctx context.Context, clientID, deviceCode string, interval, expiresIn time.Duration) (*TokenResponse, error) {
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(expiresIn)
	currentInterval := interval

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authorization cancelled: %w", ctx.Err())
		default:
		}

		if time.Until(deadline) <= 0 {
			return nil, ErrDeviceCodeExpired
		}

		tokenResp, err := pollForToken(ctx, clientID, deviceCode)
		if err == nil {
			return tokenResp, nil
		}

		switch {
		case errors.Is(err, ErrAuthorizationPending):
			// continue polling
		case errors.Is(err, ErrSlowDown):
			currentInterval += 5 * time.Second
		default:
			return nil, err
		}

		timer := time.NewTimer(currentInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("authorization cancelled: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// GitHubUser is the response from GitHub's /user endpoint.
type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// GetGitHubUser fetches the authenticated user's info.
func GetGitHubUser(ctx context.Context, token string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "entire-cli")

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is a constant
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error (%d)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}

	return &user, nil
}
