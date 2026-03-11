package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/entireio/cli/cmd/entire/cli/auth"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// Default GitHub App client ID for the device flow.
// This is a public value (no secret needed for device flow).
// Override with ENTIRE_GITHUB_CLIENT_ID env var.
const defaultGitHubClientID = "Iv23li7ashZngVIxWbpx"

func getGitHubClientID() string {
	if id := os.Getenv("ENTIRE_GITHUB_CLIENT_ID"); id != "" {
		return id
	}
	return defaultGitHubClientID
}

func newLoginCmd() *cobra.Command {
	var noBrowser bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub using device flow",
		Long: `Authenticate with GitHub to enable search and other features.

Uses GitHub's device flow: you'll get a code to enter at github.com.
The token is stored in .entire/auth.json and scoped to repo metadata access.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			clientID := getGitHubClientID()

			deviceResp, err := auth.RequestDeviceCode(ctx, clientID)
			if err != nil {
				return fmt.Errorf("requesting device code: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStdout(), "  Open this URL in your browser: %s\n", deviceResp.VerificationURI)
			fmt.Fprintf(cmd.OutOrStdout(), "  Enter code: %s\n", deviceResp.UserCode)
			fmt.Fprintln(cmd.OutOrStdout())

			if !noBrowser && deviceResp.VerificationURIComplete != "" {
				if err := browser.OpenURL(deviceResp.VerificationURIComplete); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Could not open browser automatically. Please open the URL manually.\n")
				}
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for authorization...\n")

			tokenResp, err := auth.WaitForAuthorization(
				ctx,
				clientID,
				deviceResp.DeviceCode,
				secondsToDuration(deviceResp.Interval),
				secondsToDuration(deviceResp.ExpiresIn),
			)
			if err != nil {
				return fmt.Errorf("authorization failed: %w", err)
			}

			user, err := auth.GetGitHubUser(ctx, tokenResp.AccessToken)
			if err != nil {
				return fmt.Errorf("fetching user info: %w", err)
			}

			if err := auth.SetStoredAuth(tokenResp.AccessToken, user.Login); err != nil {
				return fmt.Errorf("storing credentials: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", user.Login)
			return nil
		},
	}

	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open the browser automatically")

	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored GitHub credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := auth.DeleteStoredToken(); err != nil {
				return fmt.Errorf("removing credentials: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth-status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := auth.GetStoredToken()
			if err != nil {
				return fmt.Errorf("reading stored credentials: %w", err)
			}
			if token == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated.")
				fmt.Fprintln(cmd.OutOrStdout(), "Run 'entire login' to authenticate with GitHub.")
				return nil
			}

			var masked string
			if len(token) > 8 {
				masked = token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
			} else {
				masked = strings.Repeat("*", len(token))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Authenticated via %s\n", auth.SourceEntireDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", masked)

			if username, err := auth.GetStoredUsername(); err == nil && username != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "GitHub user: %s\n", username)
			}

			return nil
		},
	}
}

func secondsToDuration(secs int64) time.Duration {
	return time.Duration(secs) * time.Second
}
