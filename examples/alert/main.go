package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MK-Morse-SMS/Zoom-Alert"
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
	} // Create OAuth config
	config := &zoomalert.Config{
		ZoomClientID:     clientID,
		ZoomClientSecret: clientSecret,
		ZoomRedirectURI:  redirectURL,
	}

	// Create OAuth service
	oauthService := zoomalert.NewOAuthService(config)

	// Create Zoom service
	zoomService := zoomalert.NewZoomService(oauthService, robotJID, accountID)

	// Check if user is authorized
	if !zoomService.IsUserAuthorized() {
		fmt.Println("User not authorized. Please complete OAuth flow first.")
		return
	}

	// Send alert with rich content
	err := zoomService.SendAlertWithRichContent(
		targetEmail,
		"System Alert: High CPU Usage Detected",
		"ERROR",
		true,
		"This is a section block with monitoring information",
	)
	if err != nil {
		log.Fatalf("Failed to send alert: %v", err)
	}

	fmt.Println("Alert sent successfully!")
}
