package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var (
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	tenantID     = os.Getenv("TENANT_ID")
	scope        = "offline_access Tasks.ReadWrite"
	redirectURI  = "http://localhost:7861"
	serverPort   = "7861"
)

func main() {
	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("Set CLIENT_ID, CLIENT_SECRET, and TENANT_ID environment variables")
	}

	codeCh := make(chan string)

	// Start HTTP server to receive the auth code
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		code := r.FormValue("code")
		if code != "" {
			fmt.Fprintf(w, "Authorization code received. You can close this window.")
			codeCh <- code
		} else {
			fmt.Fprintf(w, "Waiting for authorization code...")
		}
	})
	go http.ListenAndServe(":"+serverPort, nil)

	// Build authorization URL
	authURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?client_id=%s&response_type=code&redirect_uri=%s&response_mode=query&scope=%s",
		tenantID, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(scope),
	)
	fmt.Println("Open the following URL in your browser to authorize:")
	fmt.Println(authURL)
	openBrowser(authURL)

	// Wait for code
	code := <-codeCh

	// Exchange code for token
	token, err := getToken(code)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}
	fmt.Println("Access token:", token.AccessToken)

	// Write credentials to file
	creds := map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"tenant_id":     tenantID,
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.ExpiresIn,
		"scope":         scope,
	}
	file, err := os.Create("oauth_credentials.json")
	if err != nil {
		log.Printf("Failed to create credentials file: %v", err)
	} else {
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		if err := enc.Encode(creds); err != nil {
			log.Printf("Failed to write credentials: %v", err)
		} else {
			fmt.Println("OAuth credentials written to oauth_credentials.json")
		}
		file.Close()
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

func getToken(code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", scope)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}
