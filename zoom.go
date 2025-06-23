package zoomalert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ZoomService handles interactions with Zoom API
type ZoomService struct {
	oauthService *OAuthService
	baseURL      string
	robotJID     string
	accountID    string
}

// User represents a Zoom user
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	JID       string `json:"jid"`
}

// UserResponse represents the response from user search
type UserResponse struct {
	Users []User `json:"users"`
}

// ChatMessage represents a chat message to be sent
type ChatMessage struct {
	RobotJID  string      `json:"robot_jid"`
	ToJID     string      `json:"to_jid"`
	AccountID string      `json:"account_id"`
	Content   ChatContent `json:"content"`
}

// ChatContent represents the content of a chat message
type ChatContent struct {
	Head     *ChatHead     `json:"head,omitempty"`
	Settings *ChatSettings `json:"settings,omitempty"`
	Body     []ChatBody    `json:"body,omitempty"`
}

// ChatHead represents the head content of a chat message (for simple text messages)
type ChatHead struct {
	Text string `json:"text"`
}

// ChatSettings represents the settings for a chat message
type ChatSettings struct {
	// Add any settings fields as needed
}

// ChatBody represents a body element in a chat message
type ChatBody struct {
	Type      string        `json:"type"`
	Layout    string        `json:"layout,omitempty"`
	Sections  []ChatSection `json:"sections,omitempty"`
	Text      string        `json:"text,omitempty"`
	Level     string        `json:"level,omitempty"`
	Closeable bool          `json:"closeable,omitempty"`
}

// ChatSection represents a section within a chat body
type ChatSection struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ChatResponse represents the response from sending a chat message
type ChatResponse struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewZoomService creates a new ZoomService
func NewZoomService(oauthService *OAuthService, robotJID, accountID string) *ZoomService {
	return &ZoomService{
		oauthService: oauthService,
		baseURL:      "https://api.zoom.us/v2",
		robotJID:     robotJID,
		accountID:    accountID,
	}
}

// GetUserByEmail gets user information using user access token (authorization code flow)
func (z *ZoomService) GetUserByEmail(email string) (*User, error) {
	token, err := z.oauthService.GetUserAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get user access token: %w", err)
	}

	// Search for user by email using user token
	url := fmt.Sprintf("%s/users/%s", z.baseURL, email)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// SendChatMessage sends a chat message using chatbot token
func (z *ZoomService) SendChatMessage(userJID, message string) error {
	token, err := z.GetChatbotToken()
	if err != nil {
		return fmt.Errorf("failed to get chatbot token: %w", err)
	}
	// Prepare chat message
	chatMsg := ChatMessage{
		RobotJID:  z.robotJID,
		ToJID:     userJID,
		AccountID: z.accountID,
		Content: ChatContent{
			Head: &ChatHead{
				Text: message,
			},
		},
	}

	jsonData, err := json.Marshal(chatMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal chat message: %w", err)
	}

	// Pretty print JSON for debugging
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		return fmt.Errorf("failed to pretty print JSON: %w", err)
	}

	// Send chat message using chatbot token
	url := fmt.Sprintf("%s/im/chat/messages", z.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log response body
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("HTTP response details (chatbot token)",
		"status", resp.Status,
		"statusCode", resp.StatusCode,
		"body", respBody.String())

	// Restore response body for potential further processing
	resp.Body = io.NopCloser(bytes.NewReader(respBody.Bytes()))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("chat message request failed with status: %d, body: %s",
			resp.StatusCode, respBody.String())
	}

	return nil
}

// SendAlertMessage sends an alert message with rich content format
func (z *ZoomService) SendAlertMessage(userJID, alertText, alertLevel string, closeable bool, sectionText string) error {
	token, err := z.GetChatbotToken()
	if err != nil {
		return fmt.Errorf("failed to get chatbot token: %w", err)
	}

	// Prepare alert message with rich content
	chatMsg := ChatMessage{
		RobotJID:  z.robotJID,
		ToJID:     userJID,
		AccountID: z.accountID,
		Content: ChatContent{
			Settings: &ChatSettings{},
			Body: []ChatBody{
				{
					Type:   "section",
					Layout: "horizontal",
					Sections: []ChatSection{
						{
							Type: "message",
							Text: sectionText,
						},
					},
				},
				{
					Type:      "alert",
					Text:      alertText,
					Level:     alertLevel,
					Closeable: closeable,
				},
			},
		},
	}

	jsonData, err := json.Marshal(chatMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal alert message: %w", err)
	}

	// Pretty print JSON for debugging
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		return fmt.Errorf("failed to pretty print JSON: %w", err)
	}
	fmt.Println("Alert Message JSON:")
	fmt.Println(prettyJSON.String())

	// Send alert message using chatbot token
	url := fmt.Sprintf("%s/im/chat/messages", z.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log response body
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("HTTP response details (alert message)",
		"status", resp.Status,
		"statusCode", resp.StatusCode,
		"body", respBody.String())

	// Restore response body for potential further processing
	resp.Body = io.NopCloser(bytes.NewReader(respBody.Bytes()))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("alert message request failed with status: %d, body: %s",
			resp.StatusCode, respBody.String())
	}

	return nil
}

// GetAuthorizationURL generates the authorization URL for OAuth flow
func (z *ZoomService) GetAuthorizationURL(state string) string {
	return z.oauthService.GetAuthorizationURL(state)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (z *ZoomService) ExchangeCodeForToken(code string) error {
	return z.oauthService.ExchangeCodeForToken(code)
}

// SendAlertWithUserToken sends alert using user authorization token (required for user lookup)
func (z *ZoomService) SendAlertWithUserToken(email, message string) error {
	// First, get the user by email using user token
	user, err := z.GetUserByEmail(email)
	if err != nil {
		slog.Error("Failed to get user with user token", "email", email, "error", err)
		return fmt.Errorf("failed to get user with user token: %w", err)
	}

	// Then send the chat message using chatbot token and user's JID
	if err := z.SendChatMessage(user.JID, message); err != nil {
		return fmt.Errorf("failed to send chat message with user token: %w", err)
	}

	return nil
}

// SendAlertWithRichContent sends alert using user authorization token with rich content format
func (z *ZoomService) SendAlertWithRichContent(email, alertText, alertLevel string, closeable bool, sectionText string) error {
	// First, get the user by email using user token
	user, err := z.GetUserByEmail(email)
	if err != nil {
		slog.Error("Failed to get user with user token", "email", email, "error", err)
		return fmt.Errorf("failed to get user with user token: %w", err)
	}

	// Then send the alert message using chatbot token and user's JID
	if err := z.SendAlertMessage(user.JID, alertText, alertLevel, closeable, sectionText); err != nil {
		return fmt.Errorf("failed to send alert message: %w", err)
	}

	return nil
}

// IsUserAuthorized checks if user authorization is available
func (z *ZoomService) IsUserAuthorized() bool {
	return z.oauthService.IsUserAuthorized()
}

// GenerateOAuthState generates a secure state parameter for OAuth flow
func (z *ZoomService) GenerateOAuthState() (string, error) {
	return z.oauthService.GenerateState()
}

// ValidateOAuthState validates and consumes an OAuth state parameter
func (z *ZoomService) ValidateOAuthState(state string) error {
	return z.oauthService.ValidateState(state)
}

// GetChatbotToken gets an access token using client credentials flow for chatbot operations
func (z *ZoomService) GetChatbotToken() (string, error) {
	// Get client credentials from oauth service's config
	config := z.oauthService.GetConfig()
	clientID := config.ZoomClientID
	clientSecret := config.ZoomClientSecret

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("client credentials not configured")
	}

	// Prepare request for client credentials flow
	url := "https://zoom.us/oauth/token?grant_type=client_credentials"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth with client credentials
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

// AlertLevel represents the severity level of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "INFO"
	AlertLevelWarning  AlertLevel = "WARNING"
	AlertLevelError    AlertLevel = "ERROR"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// CreateAlertTemplate creates a ChatContent struct for alert messages
func CreateAlertTemplate(sectionText, alertText string, level AlertLevel, closeable bool) ChatContent {
	return ChatContent{
		Settings: &ChatSettings{},
		Body: []ChatBody{
			{
				Type:   "section",
				Layout: "horizontal",
				Sections: []ChatSection{
					{
						Type: "message",
						Text: sectionText,
					},
				},
			},
			{
				Type:      "alert",
				Text:      alertText,
				Level:     string(level),
				Closeable: closeable,
			},
		},
	}
}

// SendTemplatedAlert sends an alert using a predefined template
func (z *ZoomService) SendTemplatedAlert(userJID string, content ChatContent) error {
	token, err := z.GetChatbotToken()
	if err != nil {
		return fmt.Errorf("failed to get chatbot token: %w", err)
	}

	// Prepare chat message with templated content
	chatMsg := ChatMessage{
		RobotJID:  z.robotJID,
		ToJID:     userJID,
		AccountID: z.accountID,
		Content:   content,
	}

	jsonData, err := json.Marshal(chatMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal templated alert: %w", err)
	}

	// Pretty print JSON for debugging
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		return fmt.Errorf("failed to pretty print JSON: %w", err)
	}
	fmt.Println("Templated Alert JSON:")
	fmt.Println(prettyJSON.String())

	// Send alert message using chatbot token
	url := fmt.Sprintf("%s/im/chat/messages", z.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log response body
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("HTTP response details (templated alert)",
		"status", resp.Status,
		"statusCode", resp.StatusCode,
		"body", respBody.String())

	// Restore response body for potential further processing
	resp.Body = io.NopCloser(bytes.NewReader(respBody.Bytes()))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("templated alert request failed with status: %d, body: %s",
			resp.StatusCode, respBody.String())
	}

	return nil
}
