package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	redirectURI  = getenvDefault("REDIRECT_URI", "http://localhost:7861")
	serverPort   = getenvDefault("PORT", "7861")
)

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("Set CLIENT_ID, CLIENT_SECRET, and TENANT_ID environment variables")
	}

	codeCh := make(chan string)

	// Start HTTP server to receive the auth code
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			log.Printf("ParseForm failed: %v", err)
			return
		}
		code := r.FormValue("code")
		if code != "" {
			if _, err := fmt.Fprintf(w, "Authorization code received. You can close this window."); err != nil {
				log.Printf("write response failed: %v", err)
			}
			codeCh <- code
		} else {
			if _, err := fmt.Fprintf(w, "Waiting for authorization code..."); err != nil {
				log.Printf("write response failed: %v", err)
			}
		}
	})
	go func() {
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

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
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close credentials file: %v", cerr)
		}
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
	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("closing response body failed: %v", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var terr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &terr); err == nil && (terr.Error != "" || terr.ErrorDescription != "") {
			return nil, fmt.Errorf("token endpoint %d: %s (%s)", resp.StatusCode, terr.ErrorDescription, terr.Error)
		}
		return nil, fmt.Errorf("token endpoint %d: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decoding token response failed: %w", err)
	}
	if token.AccessToken == "" {
		var terr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &terr); err == nil && (terr.Error != "" || terr.ErrorDescription != "") {
			return nil, fmt.Errorf("token response missing access_token: %s (%s)", terr.ErrorDescription, terr.Error)
		}
		return nil, fmt.Errorf("token response missing access_token; raw: %s", string(body))
	}
	return &token, nil
}
