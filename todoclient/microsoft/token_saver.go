package microsoft

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// FileTokenSaver returns a saveToken callback that persists refreshed tokens to a JSON file.
// Behavior:
// - Updates "access_token" with the latest token.AccessToken.
// - Updates "refresh_token" only if tok.RefreshToken is non-empty (Azure may not always return a new one).
// - Writes "expires_at" as RFC3339 timestamp and "expires_in" as seconds remaining (non-negative).
// - Preserves any existing fields in the JSON (client_id, client_secret, tenant_id, scope, etc.).
// - Creates the file if it does not exist.
//
// Example usage:
//   httpClient := microsoft.NewClient(msConfig, microsoft.FileTokenSaver("oauth_credentials.json"))
func FileTokenSaver(path string) func(tok *oauth2.Token) {
	return func(tok *oauth2.Token) {
		// Read existing file if present
		var creds map[string]interface{}
		if b, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(b, &creds)
		} else {
			creds = make(map[string]interface{})
		}

		// Update tokens
		creds["access_token"] = tok.AccessToken
		if tok.RefreshToken != "" {
			creds["refresh_token"] = tok.RefreshToken
		}

		// Compute expiry fields
		expiresAt := tok.Expiry
		if expiresAt.IsZero() {
			// If Expiry isn't set, do not write bogus values; skip
		} else {
			// RFC3339 absolute expiry
			creds["expires_at"] = expiresAt.UTC().Format(time.RFC3339)

			// Non-negative remaining seconds
			remaining := time.Until(expiresAt)
			if remaining < 0 {
				remaining = 0
			}
			creds["expires_in"] = int(remaining / time.Second)
		}

		// Marshal and write atomically
		tmp := path + ".tmp"
		f, err := os.Create(tmp)
		if err != nil {
			// Best-effort: if we cannot write, just return
			return
		}
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(creds); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(tmp)
			return
		}
		// Replace original
		_ = os.Rename(tmp, path)
	}
}

// FileTokenSaverWithFallback is like FileTokenSaver but ensures that if the server did not send a new refresh_token,
// the previous one from the file is preserved during write.
// Use this when your save flow also controls the "refresh_token" persistence in a single place.
// The TokenSource itself already preserves previous refresh_token in memory, so this is usually not necessary.
func FileTokenSaverWithFallback(path string) func(tok *oauth2.Token) {
	return func(tok *oauth2.Token) {
		// Read existing file if present
		var creds map[string]interface{}
		var prevRefresh string
		if b, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(b, &creds)
			if v, ok := creds["refresh_token"].(string); ok {
				prevRefresh = v
			}
		} else {
			creds = make(map[string]interface{})
		}

		// Update tokens
		creds["access_token"] = tok.AccessToken
		if tok.RefreshToken != "" {
			creds["refresh_token"] = tok.RefreshToken
		} else if prevRefresh != "" {
			creds["refresh_token"] = prevRefresh
		}

		// Compute expiry fields
		expiresAt := tok.Expiry
		if !expiresAt.IsZero() {
			creds["expires_at"] = expiresAt.UTC().Format(time.RFC3339)
			remaining := time.Until(expiresAt)
			if remaining < 0 {
				remaining = 0
			}
			creds["expires_in"] = int(remaining / time.Second)
		}

		// Write atomically
		tmp := fmt.Sprintf("%s.tmp", path)
		f, err := os.Create(tmp)
		if err != nil {
			return
		}
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(creds); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(tmp)
			return
		}
		_ = os.Rename(tmp, path)
	}
}