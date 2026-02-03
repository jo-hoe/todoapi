package microsoft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// normalizeDefaultScopes returns the default v2 scopes in stable order,
// de-duplicated and correctly cased. If provided is empty, defaults are used.
// If provided is non-empty, it will be normalized (case + dedup) and ordered
// with the default baseline first (offline_access, openid, Tasks.ReadWrite),
// with any extra scopes appended in original order skipping duplicates.
func normalizeDefaultScopes(provided []string) []string {
	// Baseline defaults in stable order
	baseline := []string{"offline_access", "openid", "Tasks.ReadWrite"}

	// Normalization function for case/canonical names
	toCanonical := func(s string) string {
		s = strings.TrimSpace(s)
		// Common variants normalization
		switch strings.ToLower(s) {
		case "offline_access":
			return "offline_access"
		case "openid":
			return "openid"
		case "tasks.readwrite", "task.readwrite", "tasksreadwrite":
			return "Tasks.ReadWrite"
		default:
			return s
		}
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(baseline))

	if len(provided) == 0 {
		// Use baseline only
		for _, b := range baseline {
			if !seen[b] {
				seen[b] = true
				result = append(result, b)
			}
		}
		return result
	}

	// First ensure baseline scopes appear first if present or needed
	for _, b := range baseline {
		c := toCanonical(b)
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}

	// Append normalized provided scopes, skipping duplicates and preserving order
	for _, p := range provided {
		c := toCanonical(p)
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}

	return result
}

// AzureV2TokenSource is a custom TokenSource for Azure AD v2.
// It ensures that the refresh_token grant includes the required scope parameter,
// and captures/persists refreshed tokens via the provided callback.
type AzureV2TokenSource struct {
	ctx       context.Context
	conf      *oauth2.Config
	http      *http.Client
	current   *oauth2.Token
	onSave    func(*oauth2.Token)
	scopesStr string
}

// NewAzureV2TokenSource constructs an AzureV2TokenSource for the given config and seed token.
// The save callback will be invoked after every successful refresh.
func NewAzureV2TokenSource(ctx context.Context, conf *oauth2.Config, seed *oauth2.Token, save func(*oauth2.Token)) oauth2.TokenSource {
	if ctx == nil {
		ctx = context.Background()
	}
	normalized := normalizeDefaultScopes(conf.Scopes)
	conf.Scopes = normalized
	scopeStr := strings.Join(normalized, " ")

	return &AzureV2TokenSource{
		ctx:       ctx,
		conf:      conf,
		http:      oauth2.NewClient(ctx, nil), // plain client without auth
		current:   seed,
		onSave:    save,
		scopesStr: scopeStr,
	}
}

// Token returns a valid token, refreshing when needed using the v2 flow with scopes.
func (ts *AzureV2TokenSource) Token() (*oauth2.Token, error) {
	// If we have a current token and it's valid, return it
	if ts.current != nil && ts.current.Valid() {
		return ts.current, nil
	}

	// If we have a refresh token, perform refresh grant with scope
	if ts.current != nil && ts.current.RefreshToken != "" {
		refreshed, err := ts.refreshWithScope(ts.current.RefreshToken)
		if err != nil {
			return nil, err
		}
		ts.current = refreshed
		if ts.onSave != nil {
			ts.onSave(refreshed)
		}
		return refreshed, nil
	}

	// Without a refresh token, we cannot refresh. Return current if set, else error.
	if ts.current != nil {
		return ts.current, nil
	}
	return nil, errors.New("no token available and no refresh_token present")
}

// refreshWithScope performs a refresh_token grant against the v2 token endpoint,
// including the required scope parameter.
func (ts *AzureV2TokenSource) refreshWithScope(refreshToken string) (*oauth2.Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("scope", ts.scopesStr)

	// Client authentication: Microsoft Identity Platform accepts client_id/client_secret in body
	if ts.conf.ClientID != "" {
		form.Set("client_id", ts.conf.ClientID)
	}
	if ts.conf.ClientSecret != "" {
		form.Set("client_secret", ts.conf.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ts.ctx, http.MethodPost, ts.conf.Endpoint.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		ExtExpiresIn int64  `json:"ext_expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if resp.StatusCode != http.StatusOK {
		var terr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&terr)
		if terr.Error != "" || terr.ErrorDescription != "" {
			return nil, fmt.Errorf("token refresh failed: %s (%s)", terr.ErrorDescription, terr.Error)
		}
		return nil, fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}
	if raw.AccessToken == "" {
		return nil, fmt.Errorf("refresh response missing access_token")
	}

	// Compute absolute expiry from expires_in
	expiry := time.Now()
	if raw.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)
	}

	// Build oauth2.Token
	out := &oauth2.Token{
		AccessToken: raw.AccessToken,
		TokenType:   raw.TokenType,
		Expiry:      expiry,
	}

	// Only persist non-empty refresh token updates
	if strings.TrimSpace(raw.RefreshToken) != "" {
		out.RefreshToken = raw.RefreshToken
	} else {
		// Keep the previous refresh token if server did not send a new one
		out.RefreshToken = refreshToken
	}

	return out, nil
}