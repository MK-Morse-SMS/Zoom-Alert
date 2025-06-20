// CLI tool for sending Zoom alerts
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kirwinrMK/Zoom-Alert"
)

func main() {
	var (
		email   = flag.String("email", "", "Recipient email address (required)")
		message = flag.String("message", "", "Message to send (required)")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		printUsage()
		return
	}

	if *email == "" || *message == "" {
		fmt.Println("Error: Both email and message are required")
		printUsage()
		os.Exit(1)
	}

	// Load configuration from environment
	config := zoomalert.LoadConfigFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize module
	module, err := zoomalert.NewZoomAlertModule(config)
	if err != nil {
		log.Fatalf("Failed to initialize ZoomAlert: %v", err)
	}

	// Check authorization
	if !module.IsUserAuthorized() {
		fmt.Println("❌ User authorization required")
		fmt.Println("Please complete OAuth authorization first:")

		authURL, err := module.GetAuthorizationURL()
		if err != nil {
			log.Fatalf("Failed to get authorization URL: %v", err)
		}

		fmt.Printf("Visit: %s\n", authURL)
		os.Exit(1)
	}

	// Send alert
	fmt.Printf("📨 Sending alert to %s...\n", *email)
	err = module.SendAlert(*email, *message)
	if err != nil {
		log.Fatalf("❌ Failed to send alert: %v", err)
	}

	fmt.Println("✅ Alert sent successfully!")
}

func printUsage() {
	fmt.Println("ZoomAlert CLI Tool")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  zoomalert-cli -email user@example.com -message 'Hello World'")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  ZOOM_ACCOUNT_ID     - Your Zoom account ID")
	fmt.Println("  ZOOM_CLIENT_ID      - Your Zoom client ID")
	fmt.Println("  ZOOM_CLIENT_SECRET  - Your Zoom client secret")
	fmt.Println("  ZOOM_ROBOT_JID      - Your Zoom robot JID")
	fmt.Println("  ZOOM_REDIRECT_URI   - OAuth redirect URI")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Send a simple alert")
	fmt.Println("  zoomalert-cli -email admin@company.com -message 'Server is down'")
	fmt.Println()
	fmt.Println("  # Send a system alert")
	fmt.Println("  zoomalert-cli -email devops@company.com -message 'Disk usage: 90%'")
}
