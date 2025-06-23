package zoomalert

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OAuthService handles Zoom OAuth authentication
type OAuthService struct {
	config           *Config
	userAccessToken  string
	userRefreshToken string
	userExpiresAt    time.Time
	// State management for OAuth flow
	stateStore map[string]StateInfo
	stateMutex sync.RWMutex
	// Token persistence
	tokenFilePath string
}

// StateInfo holds information about an OAuth state parameter
type StateInfo struct {
	CreatedAt time.Time
	ExpiresAt time.Time
}

// TokenStore represents the structure for persisting tokens
type TokenStore struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// NewOAuthService creates a new OAuthService with optional token file path
func NewOAuthService(cfg *Config, tokenFilePath ...string) *OAuthService {
	// Set default token file path if not provided
	var filePath string
	if len(tokenFilePath) > 0 && tokenFilePath[0] != "" {
		filePath = tokenFilePath[0]
	} else {
		// Default to current directory + tokens.json
		filePath = "./tokens.json"
	}

	service := &OAuthService{
		config:        cfg,
		stateStore:    make(map[string]StateInfo),
		tokenFilePath: filePath,
	}

	// Try to load existing tokens on startup
	if err := service.LoadTokens(); err != nil {
		fmt.Printf("Warning: failed to load existing tokens: %v\n", err)
	}

	return service
}

// GetAuthorizationURL generates the authorization URL for the authorization code flow
func (o *OAuthService) GetAuthorizationURL(state string) string {
	baseURL := "https://zoom.us/oauth/authorize"
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", o.config.ZoomClientID)
	params.Set("redirect_uri", o.config.ZoomRedirectURI)
	params.Set("state", state)

	return baseURL + "?" + params.Encode()
}

// ExchangeCodeForToken exchanges authorization code for access token
func (o *OAuthService) ExchangeCodeForToken(code string) error {
	if code == "" {
		return fmt.Errorf("authorization code is required")
	}

	tokenURL := "https://zoom.us/oauth/token"

	// Create the authorization header
	credentials := base64.StdEncoding.EncodeToString([]byte(o.config.ZoomClientID + ":" + o.config.ZoomClientSecret))

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", o.config.ZoomRedirectURI)

	// Create the request
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token exchange request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute token exchange request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for better error reporting
	var responseBody bytes.Buffer
	if _, err := responseBody.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OAuth code exchange failed with status %d: %s", resp.StatusCode, responseBody.String())
	}

	// Parse the response
	var tokenResp tokenResponse
	if err := json.Unmarshal(responseBody.Bytes(), &tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	// Validate response
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("no access token received in response")
	}

	// Store the user tokens
	o.userAccessToken = tokenResp.AccessToken
	o.userRefreshToken = tokenResp.RefreshToken
	o.userExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	// Auto-save tokens to file
	if err := o.SaveTokens(); err != nil {
		// Log the error but don't fail the token exchange
		fmt.Printf("Warning: failed to save tokens to file: %v\n", err)
	}

	return nil
}

// GetUserAccessToken returns a valid user access token (for authorization code flow)
func (o *OAuthService) GetUserAccessToken() (string, error) {
	// Check if we have a valid user token
	if o.userAccessToken != "" && time.Now().Before(o.userExpiresAt) {
		return o.userAccessToken, nil
	}

	// Try to refresh the user token if we have a refresh token
	if o.userRefreshToken != "" {
		return o.refreshUserToken()
	}

	return "", fmt.Errorf("no valid user access token available, authorization required")
}

// refreshUserToken refreshes the user access token using the refresh token
func (o *OAuthService) refreshUserToken() (string, error) {
	tokenURL := "https://zoom.us/oauth/token"

	// Create the authorization header
	credentials := base64.StdEncoding.EncodeToString([]byte(o.config.ZoomClientID + ":" + o.config.ZoomClientSecret))

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", o.userRefreshToken)

	// Create the request
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OAuth token refresh failed with status: %d", resp.StatusCode)
	}

	// Parse the response
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Store the refreshed user tokens
	o.userAccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		o.userRefreshToken = tokenResp.RefreshToken
	}
	o.userExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	// Auto-save refreshed tokens
	if err := o.SaveTokens(); err != nil {
		// Log the error but don't fail the token refresh
		fmt.Printf("Warning: failed to save refreshed tokens to file: %v\n", err)
	}

	return o.userAccessToken, nil
}

// GenerateState generates a secure random state parameter and stores it
func (o *OAuthService) GenerateState() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)

	// Store the state with expiration
	o.stateMutex.Lock()
	defer o.stateMutex.Unlock()

	// Clean up expired states
	o.cleanupExpiredStates()

	// Store new state (expires in 10 minutes)
	o.stateStore[state] = StateInfo{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	return state, nil
}

// ValidateState validates and consumes a state parameter
func (o *OAuthService) ValidateState(state string) error {
	if state == "" {
		return fmt.Errorf("state parameter is required")
	}

	o.stateMutex.Lock()
	defer o.stateMutex.Unlock()

	// Clean up expired states
	o.cleanupExpiredStates()

	stateInfo, exists := o.stateStore[state]
	if !exists {
		return fmt.Errorf("invalid or expired state parameter")
	}

	// Check if state has expired
	if time.Now().After(stateInfo.ExpiresAt) {
		delete(o.stateStore, state)
		return fmt.Errorf("state parameter has expired")
	}

	// Consume the state (remove it so it can't be reused)
	delete(o.stateStore, state)

	return nil
}

// cleanupExpiredStates removes expired state entries (must be called with mutex held)
func (o *OAuthService) cleanupExpiredStates() {
	now := time.Now()
	for state, info := range o.stateStore {
		if now.After(info.ExpiresAt) {
			delete(o.stateStore, state)
		}
	}
}

// IsUserAuthorized checks if we have a valid user access token
func (o *OAuthService) IsUserAuthorized() bool {
	_, err := o.GetUserAccessToken()
	return err == nil
}

// GetConfig returns the OAuth configuration (for internal use by other services)
func (o *OAuthService) GetConfig() *Config {
	return o.config
}

// SaveTokens saves tokens to the configured file path
func (o *OAuthService) SaveTokens() error {
	if o.tokenFilePath == "" {
		return fmt.Errorf("no token file path configured")
	}

	store := TokenStore{
		AccessToken:  o.userAccessToken,
		RefreshToken: o.userRefreshToken,
		ExpiresAt:    o.userExpiresAt,
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(o.tokenFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(o.tokenFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadTokens loads tokens from the configured file path
func (o *OAuthService) LoadTokens() error {
	if o.tokenFilePath == "" {
		return fmt.Errorf("no token file path configured")
	}

	data, err := os.ReadFile(o.tokenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Token file doesn't exist yet, this is normal for first run
			return nil
		}
		return fmt.Errorf("failed to read token file: %w", err)
	}

	var store TokenStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	// Only load tokens if they haven't expired
	if time.Now().Before(store.ExpiresAt) {
		o.userAccessToken = store.AccessToken
		o.userRefreshToken = store.RefreshToken
		o.userExpiresAt = store.ExpiresAt
	}

	return nil
}

// GetTokenFilePath returns the configured token file path
func (o *OAuthService) GetTokenFilePath() string {
	return o.tokenFilePath
}

// SetTokenFilePath updates the token file path
func (o *OAuthService) SetTokenFilePath(path string) {
	o.tokenFilePath = path
}
