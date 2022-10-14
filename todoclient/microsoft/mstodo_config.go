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
		Scopes:       []string{"openid offline_access tasks.readwrite"},
		RedirectURL:  "https://localhost/login/authorized",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	var token = oauth2.Token{
		AccessToken:  config.Token.AccessToken,
		TokenType:    config.Token.TokenType,
		RefreshToken: config.Token.RefreshToken,
		Expiry:       time.Now(),
	}
	ctx := context.Background()
	tokenSource := oauthConfig.TokenSource(ctx, &token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln(err)
	}

	// access-token is valid for 1 hour (by default)
	// and the issued refresh token is valid for 90 days
	// ms sends a new refresh token for each call back which again is 90 days valid
	if saveToken != nil {
		saveToken(newToken)
	}

	return oauth2.NewClient(ctx, tokenSource)
}
