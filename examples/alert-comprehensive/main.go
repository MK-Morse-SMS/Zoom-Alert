package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kirwinrMK/Zoom-Alert"
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

	// Create OAuth config
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

	// Get user information
	user, err := zoomService.GetUserByEmail(targetEmail)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("Sending alerts to user: %s (%s)\n", user.Email, user.JID)

	// Example 1: Send a simple alert with rich content
	fmt.Println("\n1. Sending ERROR alert...")
	err = zoomService.SendAlertWithRichContent(
		targetEmail,
		"System Alert: High CPU Usage Detected (95%)",
		"ERROR",
		true,
		"Server monitoring has detected high resource usage on production server.",
	)
	if err != nil {
		log.Printf("Failed to send ERROR alert: %v", err)
	} else {
		fmt.Println("ERROR alert sent successfully!")
	}

	// Example 2: Send a WARNING alert using the template function
	fmt.Println("\n2. Sending WARNING alert using template...")
	warningContent := zoomalert.CreateAlertTemplate(
		"Database monitoring system has detected slow query performance.",
		"Warning: Database query response time is above threshold (2.5s average)",
		zoomalert.AlertLevelWarning,
		true,
	)

	err = zoomService.SendTemplatedAlert(user.JID, warningContent)
	if err != nil {
		log.Printf("Failed to send WARNING alert: %v", err)
	} else {
		fmt.Println("WARNING alert sent successfully!")
	}

	// Example 3: Send an INFO alert
	fmt.Println("\n3. Sending INFO alert...")
	infoContent := zoomalert.CreateAlertTemplate(
		"Scheduled maintenance notification for tonight's deployment.",
		"Info: Maintenance window scheduled for 2:00 AM - 4:00 AM EST",
		zoomalert.AlertLevelInfo,
		false, // Non-closeable for important info
	)

	err = zoomService.SendTemplatedAlert(user.JID, infoContent)
	if err != nil {
		log.Printf("Failed to send INFO alert: %v", err)
	} else {
		fmt.Println("INFO alert sent successfully!")
	}

	// Example 4: Send a CRITICAL alert
	fmt.Println("\n4. Sending CRITICAL alert...")
	criticalContent := zoomalert.CreateAlertTemplate(
		"URGENT: Production system failure detected. Immediate action required.",
		"CRITICAL: Application server is down - users cannot access the system",
		zoomalert.AlertLevelCritical,
		false, // Non-closeable for critical alerts
	)

	err = zoomService.SendTemplatedAlert(user.JID, criticalContent)
	if err != nil {
		log.Printf("Failed to send CRITICAL alert: %v", err)
	} else {
		fmt.Println("CRITICAL alert sent successfully!")
	}

	fmt.Println("\nAll alerts have been processed!")
}
