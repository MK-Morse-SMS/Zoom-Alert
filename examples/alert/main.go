package main

import (
	"fmt"
	"log"
	"os"

	zoomalert "github.com/MK-Morse-SMS/Zoom-Alert"
)

func main() {
	// Get environment variables
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")
	redirectURL := os.Getenv("ZOOM_REDIRECT_URL")
	robotJID := os.Getenv("ZOOM_ROBOT_JID")
	accountID := os.Getenv("ZOOM_ACCOUNT_ID")
	targetEmail := os.Getenv("TARGET_EMAIL")

	if clientID == "" || clientSecret == "" || redirectURL == "" || robotJID == "" || accountID == "" || targetEmail == "" {
		log.Fatal("Missing required environment variables: ZOOM_CLIENT_ID, ZOOM_CLIENT_SECRET, ZOOM_REDIRECT_URL, ZOOM_ROBOT_JID, ZOOM_ACCOUNT_ID, TARGET_EMAIL")
	}

	// Create configuration with custom token file path
	config := &zoomalert.Config{
		ZoomClientID:     clientID,
		ZoomClientSecret: clientSecret,
		ZoomRedirectURI:  redirectURL,
		ZoomRobotJID:     robotJID,
		ZoomAccountID:    accountID,
		TokenFilePath:    "./zoom_tokens.json", // Custom token file location
	}

	// Initialize the ZoomAlert module
	module, err := zoomalert.NewZoomAlertModule(config)
	if err != nil {
		log.Fatalf("Failed to initialize ZoomAlert module: %v", err)
	}

	// Check if user is authorized
	if !module.IsUserAuthorized() {
		fmt.Println("User not authorized. Please complete OAuth flow first.")
		fmt.Printf("Visit: /api/v1/oauth/authorize to start the OAuth flow\n")
		return
	}

	// Send alert using the module
	err = module.SendAlert(targetEmail, "System Alert: High CPU Usage Detected - This is a test alert!")
	if err != nil {
		log.Fatalf("Failed to send alert: %v", err)
	}

	fmt.Println("Alert sent successfully!")
	fmt.Printf("Tokens are persisted in: %s\n", config.TokenFilePath)
}
