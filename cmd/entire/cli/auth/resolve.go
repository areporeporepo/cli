package auth

// SourceEntireDir is the display name for the device flow token source.
const SourceEntireDir = ".entire/auth.json"

// ResolveGitHubToken returns the token stored by the device flow login.
// If no token is found, an empty string is returned with no error.
func ResolveGitHubToken() (string, error) {
	token, err := GetStoredToken()
	if err != nil {
		return "", err
	}
	return token, nil
}
