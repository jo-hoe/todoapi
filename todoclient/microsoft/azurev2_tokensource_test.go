package microsoft

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNormalizeDefaultScopes(t *testing.T) {
	t.Run("defaults when none provided", func(t *testing.T) {
		out := normalizeDefaultScopes(nil)
		want := []string{"offline_access", "openid", "Tasks.ReadWrite"}
		if len(out) != len(want) {
			t.Fatalf("expected %d scopes, got %d (%v)", len(want), len(out), out)
		}
		for i := range want {
			if out[i] != want[i] {
				t.Errorf("expected scope[%d]=%q, got %q", i, want[i], out[i])
			}
		}
	})
	t.Run("normalizes case and de-duplicates, keeps stable order", func(t *testing.T) {
		in := []string{"tasks.readwrite", "offline_access", "openid", "Tasks.ReadWrite", "extra.scope"}
		out := normalizeDefaultScopes(in)
		// Baseline first, then extra unique in input order
		want := []string{"offline_access", "openid", "Tasks.ReadWrite", "extra.scope"}
		if len(out) != len(want) {
			t.Fatalf("expected %d scopes, got %d (%v)", len(want), len(out), out)
		}
		for i := range want {
			if out[i] != want[i] {
				t.Errorf("expected scope[%d]=%q, got %q", i, want[i], out[i])
			}
		}
	})
}

func TestAzureV2TokenSource_RefreshIncludesScopeAndPersists(t *testing.T) {
	// Capture the received request body for assertions
	var capturedBody string

	// Mock v2 token endpoint
	tsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)

		// Parse captured body as application/x-www-form-urlencoded instead of re-reading r.Body
		formVals, _ := url.ParseQuery(capturedBody)
		if formVals.Get("grant_type") != "refresh_token" {
			http.Error(w, "missing or wrong grant_type", http.StatusBadRequest)
			return
		}
		if formVals.Get("refresh_token") == "" {
			http.Error(w, "missing refresh_token", http.StatusBadRequest)
			return
		}
		scope := formVals.Get("scope")
		if scope == "" {
			http.Error(w, "missing scope", http.StatusBadRequest)
			return
		}
		// Ensure expected normalized default scopes present (space-delimited)
		// Order must be: offline_access openid Tasks.ReadWrite
		if scope != "offline_access openid Tasks.ReadWrite" {
			http.Error(w, "unexpected scope order/content: "+scope, http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"access_token":   "new_access",
			"token_type":     "Bearer",
			"expires_in":     3600,
			"ext_expires_in": 7200,
			"refresh_token":  "new_refresh",
			"scope":          scope,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tsrv.Close()

	// Build oauth2.Config targeting our mock server
	conf := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       nil, // let TokenSource normalize to defaults
		RedirectURL:  "https://localhost/login/authorized",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: tsrv.URL, // point to our mock
		},
	}

	// Seed token: expired, with a refresh_token
	seed := &oauth2.Token{
		AccessToken:  "old_access",
		TokenType:    "Bearer",
		RefreshToken: "old_refresh",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	// Track save callback invocation
	var saved *oauth2.Token
	save := func(tok *oauth2.Token) {
		saved = tok
	}

	ts := NewAzureV2TokenSource(context.Background(), conf, seed, save)

	// Invoke Token to trigger refresh
	got, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}
	if got == nil {
		t.Fatal("Token() returned nil token")
	}
	if got.AccessToken != "new_access" {
		t.Errorf("expected access_token=new_access, got %s", got.AccessToken)
	}
	if got.RefreshToken != "new_refresh" {
		t.Errorf("expected refresh_token=new_refresh, got %s", got.RefreshToken)
	}
	if saved == nil {
		t.Fatal("save callback was not invoked")
	}
	if saved.AccessToken != "new_access" {
		t.Errorf("expected saved access_token=new_access, got %s", saved.AccessToken)
	}

	// Assert raw body contains scope parameter as space-delimited string
	formVals, err := url.ParseQuery(capturedBody)
	if err != nil {
		t.Fatalf("failed to parse captured body: %v", err)
	}
	if formVals.Get("scope") != "offline_access openid Tasks.ReadWrite" {
		t.Errorf("expected scope 'offline_access openid Tasks.ReadWrite', got %q", formVals.Get("scope"))
	}

	// Now test the case where server does NOT return a new refresh_token
	capturedBody = ""
	tsrvNoRT := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = r.ParseForm()
		resp := map[string]interface{}{
			"access_token": "another_access",
			"token_type":   "Bearer",
			"expires_in":   1800,
			"scope":        r.Form.Get("scope"),
			// refresh_token intentionally omitted/empty
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tsrvNoRT.Close()

	conf.Endpoint.TokenURL = tsrvNoRT.URL
	// Use token source with prior refreshed token
	ts2 := NewAzureV2TokenSource(context.Background(), conf, got, nil)

	got2, err := ts2.Token()
	if err != nil {
		t.Fatalf("Token() error (no RT returned): %v", err)
	}
	if got2.RefreshToken != got.RefreshToken {
		t.Errorf("expected refresh_token to persist previous (%s), got %s", got.RefreshToken, got2.RefreshToken)
	}
}

func TestAzureV2TokenSource_UsesProvidedScopesNormalized(t *testing.T) {
	// Provided scopes missing openid and with varied casing/duplicates
	provided := []string{"tasks.readwrite", "offline_access", "Tasks.ReadWrite"}

	// Echo handler to verify scope string matches normalization result
	var seenScope string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		seenScope = r.Form.Get("scope")
		resp := map[string]interface{}{
			"access_token": "x",
			"token_type":   "Bearer",
			"expires_in":   10,
			"scope":        seenScope,
			"refresh_token": "rt",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	conf := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       provided,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: srv.URL,
		},
	}

	seed := &oauth2.Token{
		AccessToken:  "old",
		TokenType:    "Bearer",
		RefreshToken: "rt",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	ts := NewAzureV2TokenSource(context.Background(), conf, seed, nil)
	_, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}

	// Expected normalized scopes: baseline then extras unique
	want := "offline_access openid Tasks.ReadWrite"
	if !strings.HasPrefix(seenScope, want) {
		t.Errorf("expected scope to start with %q, got %q", want, seenScope)
	}
}