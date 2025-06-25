// Package zoomalert provides a Zoom Alert Service module for sending messages to Zoom users
package zoomalert

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Option func(*ZoomAlertModule)

// WithLogger sets a custom logger for the ZoomAlertModule
func WithLogger(logger *slog.Logger) Option {
	return func(m *ZoomAlertModule) {
		m.logger = logger
	}
}

// ZoomAlertModule represents the main module that can be integrated into other projects
type ZoomAlertModule struct {
	config       *Config
	oauthService *OAuthService
	zoomService  *ZoomService
	server       *http.Server
	logger       *slog.Logger
}

// Config holds the configuration for the Zoom Alert Service
type Config struct {
	ZoomAccountID    string
	ZoomClientID     string
	ZoomClientSecret string
	ZoomRedirectURI  string
	ZoomRobotJID     string
	Port             string
	TokenFilePath    string
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Port:          "8080",
		TokenFilePath: "./tokens.json",
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}
	config := DefaultConfig()

	if val := os.Getenv("ZOOM_ACCOUNT_ID"); val != "" {
		config.ZoomAccountID = val
	}
	if val := os.Getenv("ZOOM_CLIENT_ID"); val != "" {
		config.ZoomClientID = val
	}
	if val := os.Getenv("ZOOM_CLIENT_SECRET"); val != "" {
		config.ZoomClientSecret = val
	}
	if val := os.Getenv("ZOOM_REDIRECT_URI"); val != "" {
		config.ZoomRedirectURI = val
	}
	if val := os.Getenv("ZOOM_ROBOT_JID"); val != "" {
		config.ZoomRobotJID = val
	}
	if val := os.Getenv("PORT"); val != "" {
		config.Port = val
	}
	if val := os.Getenv("TOKEN_FILE_PATH"); val != "" {
		config.TokenFilePath = val
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ZoomAccountID == "" {
		return fmt.Errorf("ZOOM_ACCOUNT_ID is required")
	}
	if c.ZoomClientID == "" {
		return fmt.Errorf("ZOOM_CLIENT_ID is required")
	}
	if c.ZoomClientSecret == "" {
		return fmt.Errorf("ZOOM_CLIENT_SECRET is required")
	}
	return nil
}

// NewZoomAlertModule creates a new ZoomAlertModule with the given configuration
func NewZoomAlertModule(config *Config, options ...Option) (*ZoomAlertModule, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	ms := &ZoomAlertModule{
		config: config,
		logger: slog.Default(),
	}
	// Apply options
	for _, opt := range options {
		opt(ms)
	}

	// Initialize services
	ms.oauthService = NewOAuthService(config, ms.logger, config.TokenFilePath)
	ms.zoomService = NewZoomService(ms.oauthService, config.ZoomRobotJID, config.ZoomAccountID, ms.logger)

	return ms, nil
}

// SendMessage sends a message to a Zoom user by email
func (m *ZoomAlertModule) SendMessage(email string, message ZoomContent) error {
	if !m.zoomService.IsUserAuthorized() {
		return fmt.Errorf("user is not authorized")
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if err := m.zoomService.SendMessageByEmail(email, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	m.logger.Info("Message sent successfully", "email", email)
	return nil
}

// IsUserAuthorized checks if the module has user authorization
func (m *ZoomAlertModule) IsUserAuthorized() bool {
	return m.zoomService.IsUserAuthorized()
}

// GetAuthorizationURL returns the OAuth authorization URL
func (m *ZoomAlertModule) GetAuthorizationURL() (string, error) {
	state, err := m.oauthService.GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	url := m.oauthService.GetAuthorizationURL(state)
	return url, nil
}

// HandleOAuthCallback processes the OAuth callback
func (m *ZoomAlertModule) HandleOAuthCallback(code, state string) error {
	if err := m.oauthService.ValidateState(state); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	return m.oauthService.ExchangeCodeForToken(code)
}

// Shutdown gracefully shuts down the HTTP server
func (m *ZoomAlertModule) Shutdown() error {
	if m.server == nil {
		return nil
	}

	m.logger.Info("Shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.server.Shutdown(ctx)
}

// RegisterOAuthRoutes sets up the OAuth routes on an existing Gin router
func (m *ZoomAlertModule) RegisterOAuthRoutes(router *gin.Engine) {
	alertHandler := NewAlertHandler(m.zoomService)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", alertHandler.HealthCheck)
		v1.GET("/auth/status", alertHandler.GetAuthStatus)
		v1.GET("/oauth/callback", alertHandler.OAuthCallback)
		v1.GET("/oauth/authorize", alertHandler.OAuthAuthorize)
	}
}

// GetZoomService returns the underlying ZoomService for advanced usage
func (m *ZoomAlertModule) GetZoomService() *ZoomService {
	return m.zoomService
}

// GetOAuthService returns the underlying OAuthService for advanced usage
func (m *ZoomAlertModule) GetOAuthService() *OAuthService {
	return m.oauthService
}

// Logger returns the module's logger
func (m *ZoomAlertModule) Logger() *slog.Logger {
	return m.logger
}
