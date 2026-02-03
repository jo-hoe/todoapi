package microsoft

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type MSClientConfig struct {
	ClientCredentials MSClientCredentials
	Token             MsOAuthToken
	Scopes           []string
}

type MSClientCredentials struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type MsOAuthToken struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewClient(config MSClientConfig, saveToken func(token *oauth2.Token)) *http.Client {
	var oauthConfig = oauth2.Config{
		ClientID:     config.ClientCredentials.ClientId,
		ClientSecret: config.ClientCredentials.ClientSecret,
		// Use provided scopes if any; normalization and defaulting will be handled by the AzureV2TokenSource
		Scopes:      config.Scopes,
		RedirectURL: "https://localhost/login/authorized",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	var token = oauth2.Token{
		AccessToken:  config.Token.AccessToken,
		TokenType:    config.Token.TokenType,
		RefreshToken: config.Token.RefreshToken,
		// Force immediate refresh on startup to ensure we persist a fresh token and correct scopes
		Expiry: time.Now().Add(-1 * time.Hour),
	}
	ctx := context.Background()

	// Use custom Azure AD v2 TokenSource to ensure scope is included on refresh and tokens are persisted
	ts := NewAzureV2TokenSource(ctx, &oauthConfig, &token, saveToken)
	if _, err := ts.Token(); err != nil {
		log.Fatalln(err)
	}

	return oauth2.NewClient(ctx, ts)
}
