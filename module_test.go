package zoomalert

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ZoomAccountID:    "test_account_id",
				ZoomClientID:     "test_client_id",
				ZoomClientSecret: "test_client_secret",
			},
			wantErr: false,
		},
		{
			name: "missing account id",
			config: &Config{
				ZoomClientID:     "test_client_id",
				ZoomClientSecret: "test_client_secret",
			},
			wantErr: true,
		},
		{
			name: "missing client id",
			config: &Config{
				ZoomAccountID:    "test_account_id",
				ZoomClientSecret: "test_client_secret",
			},
			wantErr: true,
		},
		{
			name: "missing client secret",
			config: &Config{
				ZoomAccountID: "test_account_id",
				ZoomClientID:  "test_client_id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Port)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.LogLevel)
	}
}

func TestNewZoomAlertModule(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
		ZoomRobotJID:     "test_robot_jid",
		Port:             "8080",
		LogLevel:         "info",
	}

	module, err := NewZoomAlertModule(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if module == nil {
		t.Fatal("Expected module to be created")
	}

	if module.config != config {
		t.Error("Expected config to be set")
	}

	if module.oauthService == nil {
		t.Error("Expected oauth service to be initialized")
	}

	if module.zoomService == nil {
		t.Error("Expected zoom service to be initialized")
	}

	if module.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestOAuthService_GenerateState(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)

	state, err := oauth.GenerateState()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if state == "" {
		t.Error("Expected state to be generated")
	}

	// Check if state is stored
	oauth.stateMutex.RLock()
	_, exists := oauth.stateStore[state]
	oauth.stateMutex.RUnlock()

	if !exists {
		t.Error("Expected state to be stored")
	}
}

func TestOAuthService_ValidateState(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)

	// Generate a state
	state, err := oauth.GenerateState()
	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	// Validate the state
	err = oauth.ValidateState(state)
	if err != nil {
		t.Errorf("Expected state validation to pass, got %v", err)
	}

	// Try to validate the same state again (should fail as it's consumed)
	err = oauth.ValidateState(state)
	if err == nil {
		t.Error("Expected state validation to fail on second use")
	}
}

func TestOAuthService_ValidateExpiredState(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)

	// Manually add an expired state
	expiredState := "expired_state"
	oauth.stateMutex.Lock()
	oauth.stateStore[expiredState] = StateInfo{
		CreatedAt: time.Now().Add(-15 * time.Minute),
		ExpiresAt: time.Now().Add(-5 * time.Minute),
	}
	oauth.stateMutex.Unlock()

	// Try to validate expired state
	err := oauth.ValidateState(expiredState)
	if err == nil {
		t.Error("Expected expired state validation to fail")
	}
}

func TestOAuthService_IsUserAuthorized(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)

	// Should not be authorized initially
	if oauth.IsUserAuthorized() {
		t.Error("Expected user to not be authorized initially")
	}

	// Set a valid token
	oauth.userAccessToken = "test_token"
	oauth.userExpiresAt = time.Now().Add(1 * time.Hour)

	// Should be authorized now
	if !oauth.IsUserAuthorized() {
		t.Error("Expected user to be authorized with valid token")
	}

	// Set expired token
	oauth.userExpiresAt = time.Now().Add(-1 * time.Hour)

	// Should not be authorized with expired token
	if oauth.IsUserAuthorized() {
		t.Error("Expected user to not be authorized with expired token")
	}
}

func TestNewZoomService(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)
	zoom := NewZoomService(oauth, "test_robot_jid", "test_account_id")

	if zoom == nil {
		t.Fatal("Expected zoom service to be created")
	}

	if zoom.oauthService != oauth {
		t.Error("Expected oauth service to be set")
	}

	if zoom.robotJID != "test_robot_jid" {
		t.Error("Expected robot JID to be set")
	}

	if zoom.accountID != "test_account_id" {
		t.Error("Expected account ID to be set")
	}

	if zoom.baseURL != "https://api.zoom.us/v2" {
		t.Error("Expected base URL to be set correctly")
	}
}

func TestZoomService_IsUserAuthorized(t *testing.T) {
	config := &Config{
		ZoomAccountID:    "test_account_id",
		ZoomClientID:     "test_client_id",
		ZoomClientSecret: "test_client_secret",
	}

	oauth := NewOAuthService(config)
	zoom := NewZoomService(oauth, "test_robot_jid", "test_account_id")

	// Should not be authorized initially
	if zoom.IsUserAuthorized() {
		t.Error("Expected user to not be authorized initially")
	}

	// Set a valid token in oauth service
	oauth.userAccessToken = "test_token"
	oauth.userExpiresAt = time.Now().Add(1 * time.Hour)

	// Should be authorized now
	if !zoom.IsUserAuthorized() {
		t.Error("Expected user to be authorized with valid token")
	}
}
