package zoomalert

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AlertHandler handles HTTP requests for alert operations
type AlertHandler struct {
	zoomService *ZoomService
}

// AlertRequest represents the request payload for sending alerts
type AlertRequest struct {
	Email   string `json:"email" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// AlertResponse represents the response from alert operations
type AlertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewAlertHandler creates a new AlertHandler
func NewAlertHandler(zoomService *ZoomService) *AlertHandler {
	return &AlertHandler{
		zoomService: zoomService,
	}
}

// SendAlert sends alert using the best available authorization method
func (h *AlertHandler) SendAlert(c *gin.Context) {
	if !h.zoomService.IsUserAuthorized() {
		c.JSON(http.StatusUnauthorized, AlertResponse{
			Success: false,
			Message: "User is not authorized",
		})
		return
	}

	var req AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Validate email and message
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Message is required",
		})
		return
	}

	err := h.zoomService.SendAlertWithUserToken(req.Email, req.Message)
	if err != nil {
		slog.Error("Failed to send alert with authorization:", "error", err)
		c.JSON(http.StatusInternalServerError, AlertResponse{
			Success: false,
			Message: "Failed to send alert",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AlertResponse{
		Success: true,
		Message: "Alert sent successfully",
	})
}

// HealthCheck returns the health status of the service
func (h *AlertHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "zoom-alert-service",
	})
}

// OAuthAuthorize initiates the OAuth authorization flow
func (h *AlertHandler) OAuthAuthorize(c *gin.Context) {
	// Generate a secure state parameter for CSRF protection
	state, err := h.zoomService.GenerateOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate OAuth state: " + err.Error(),
		})
		return
	}

	// Get the authorization URL
	authURL := h.zoomService.GetAuthorizationURL(state)

	// Return both the URL and state for the frontend
	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":             state,
		"message":           "Please visit the authorization URL to complete OAuth flow",
	})
}

// OAuthCallback handles the OAuth callback
func (h *AlertHandler) OAuthCallback(c *gin.Context) {
	// Extract parameters
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDescription := c.Query("error_description")

	// Handle OAuth errors
	if errorParam != "" {
		errorMsg := "OAuth authorization failed: " + errorParam
		if errorDescription != "" {
			errorMsg += " (" + errorDescription + ")"
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errorParam,
			"message": errorMsg,
		})
		return
	}

	// Validate required parameters
	if code == "" {
		errorMsg := "Missing authorization code in callback"
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg,
		})
		return
	}

	// Validate state parameter for CSRF protection
	if err := h.zoomService.ValidateOAuthState(state); err != nil {
		errorMsg := "Invalid or expired state parameter: " + err.Error()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg,
		})
		return
	}

	// Exchange code for token
	if err := h.zoomService.ExchangeCodeForToken(code); err != nil {
		errorMsg := "Failed to exchange code for token: " + err.Error()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errorMsg,
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"message": "Authorization successful",
		"status":  "authorized",
	})
}

// GetAuthStatus returns the current authorization status
func (h *AlertHandler) GetAuthStatus(c *gin.Context) {
	isAuthorized := h.zoomService.IsUserAuthorized()

	c.JSON(http.StatusOK, gin.H{
		"user_authorized": isAuthorized,
		"message": func() string {
			if isAuthorized {
				return "User authorization available - full user lookup enabled"
			}
			return "Only server-to-server authorization available - limited functionality"
		}(),
	})
}

// RichAlertRequest represents the request payload for sending rich formatted alerts
type RichAlertRequest struct {
	Email       string `json:"email" binding:"required"`
	AlertText   string `json:"alert_text" binding:"required"`
	AlertLevel  string `json:"alert_level" binding:"required"`
	SectionText string `json:"section_text" binding:"required"`
	Closeable   bool   `json:"closeable"`
}

// TemplatedAlertRequest represents the request payload for sending templated alerts
type TemplatedAlertRequest struct {
	Email       string `json:"email" binding:"required"`
	AlertText   string `json:"alert_text" binding:"required"`
	AlertLevel  string `json:"alert_level" binding:"required"`
	SectionText string `json:"section_text" binding:"required"`
	Closeable   bool   `json:"closeable"`
}

// SendRichAlert sends a rich formatted alert with sections and alert blocks
func (h *AlertHandler) SendRichAlert(c *gin.Context) {
	if !h.zoomService.IsUserAuthorized() {
		c.JSON(http.StatusUnauthorized, AlertResponse{
			Success: false,
			Message: "User is not authorized",
		})
		return
	}

	var req RichAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	if req.AlertText == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Alert text is required",
		})
		return
	}

	if req.SectionText == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Section text is required",
		})
		return
	}

	// Validate alert level
	validLevels := map[string]bool{
		"INFO":     true,
		"WARNING":  true,
		"ERROR":    true,
		"CRITICAL": true,
	}
	if !validLevels[req.AlertLevel] {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Invalid alert level. Must be one of: INFO, WARNING, ERROR, CRITICAL",
		})
		return
	}

	// Send rich alert
	err := h.zoomService.SendAlertWithRichContent(
		req.Email,
		req.AlertText,
		req.AlertLevel,
		req.Closeable,
		req.SectionText,
	)
	if err != nil {
		slog.Error("Failed to send rich alert:", "error", err)
		c.JSON(http.StatusInternalServerError, AlertResponse{
			Success: false,
			Message: "Failed to send rich alert",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AlertResponse{
		Success: true,
		Message: "Rich alert sent successfully",
	})
}

// SendTemplatedAlert sends an alert using the template system
func (h *AlertHandler) SendTemplatedAlert(c *gin.Context) {
	if !h.zoomService.IsUserAuthorized() {
		c.JSON(http.StatusUnauthorized, AlertResponse{
			Success: false,
			Message: "User is not authorized",
		})
		return
	}

	var req TemplatedAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	// Get user first
	user, err := h.zoomService.GetUserByEmail(req.Email)
	if err != nil {
		slog.Error("Failed to get user:", "email", req.Email, "error", err)
		c.JSON(http.StatusNotFound, AlertResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}
	// Create alert template
	var alertLevel AlertLevel
	switch req.AlertLevel {
	case "INFO":
		alertLevel = AlertLevelInfo
	case "WARNING":
		alertLevel = AlertLevelWarning
	case "ERROR":
		alertLevel = AlertLevelError
	case "CRITICAL":
		alertLevel = AlertLevelCritical
	default:
		c.JSON(http.StatusBadRequest, AlertResponse{
			Success: false,
			Message: "Invalid alert level. Must be one of: INFO, WARNING, ERROR, CRITICAL",
		})
		return
	}

	content := CreateAlertTemplate(
		req.SectionText,
		req.AlertText,
		alertLevel,
		req.Closeable,
	)

	// Send templated alert
	err = h.zoomService.SendTemplatedAlert(user.JID, content)
	if err != nil {
		slog.Error("Failed to send templated alert:", "error", err)
		c.JSON(http.StatusInternalServerError, AlertResponse{
			Success: false,
			Message: "Failed to send templated alert",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AlertResponse{
		Success: true,
		Message: "Templated alert sent successfully",
	})
}
