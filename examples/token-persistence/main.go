package main

import (
	"fmt"
	"log"
	"os"

	zoomalert "github.com/MK-Morse-SMS/Zoom-Alert"
)

func main() {
	fmt.Println("🔐 ZoomAlert Token Persistence Demo")
	fmt.Println("===================================")

	// Example 1: Using default token file path
	fmt.Println("\n1. Using default token file path (./tokens.json)")
	config1 := zoomalert.LoadConfigFromEnv()
	fmt.Printf("   Default token file path: %s\n", config1.TokenFilePath)

	// Example 2: Using custom token file path in config
	fmt.Println("\n2. Using custom token file path via config")
	config2 := &zoomalert.Config{
		ZoomClientID:     os.Getenv("ZOOM_CLIENT_ID"),
		ZoomClientSecret: os.Getenv("ZOOM_CLIENT_SECRET"),
		ZoomRedirectURI:  os.Getenv("ZOOM_REDIRECT_URL"),
		ZoomRobotJID:     os.Getenv("ZOOM_ROBOT_JID"),
		ZoomAccountID:    os.Getenv("ZOOM_ACCOUNT_ID"),
		TokenFilePath:    "/app/data/zoom_tokens.json", // Custom path for Docker
		Port:             "8080",
		LogLevel:         "info",
	}
	fmt.Printf("   Custom token file path: %s\n", config2.TokenFilePath)

	// Example 3: Using environment variable for token file path
	fmt.Println("\n3. Using environment variable TOKEN_FILE_PATH")
	os.Setenv("TOKEN_FILE_PATH", "/tmp/my_zoom_tokens.json")
	config3 := zoomalert.LoadConfigFromEnv()
	fmt.Printf("   Environment-based token file path: %s\n", config3.TokenFilePath)

	// Example 4: Initialize module and show token persistence
	fmt.Println("\n4. Initializing module with token persistence")
	if config1.ZoomClientID == "" {
		fmt.Println("   ⚠️  No ZOOM_CLIENT_ID found in environment")
		fmt.Println("   Set environment variables to test token persistence:")
		fmt.Println("   - ZOOM_CLIENT_ID")
		fmt.Println("   - ZOOM_CLIENT_SECRET")
		fmt.Println("   - ZOOM_REDIRECT_URL")
		fmt.Println("   - ZOOM_ROBOT_JID")
		fmt.Println("   - ZOOM_ACCOUNT_ID")
		return
	}

	module, err := zoomalert.NewZoomAlertModule(config1)
	if err != nil {
		log.Printf("   ⚠️  Failed to initialize module: %v", err)
		return
	}

	fmt.Printf("   ✅ Module initialized successfully\n")
	fmt.Printf("   📁 Token file location: %s\n", config1.TokenFilePath)

	// Check if tokens exist
	if _, err := os.Stat(config1.TokenFilePath); err == nil {
		fmt.Printf("   💾 Existing token file found - tokens will be loaded automatically\n")
	} else {
		fmt.Printf("   🆕 No existing token file - will be created after OAuth flow\n")
	}

	// Show authorization status
	if module.IsUserAuthorized() {
		fmt.Printf("   ✅ User is authorized (tokens loaded from file)\n")
	} else {
		fmt.Printf("   ⚠️  User needs authorization - visit OAuth URL to authenticate\n")
	}

	fmt.Println("\n💡 Token Persistence Benefits:")
	fmt.Println("   - Tokens survive application restarts")
	fmt.Println("   - Automatic token refresh when expired")
	fmt.Println("   - Configurable storage location")
	fmt.Println("   - Docker-friendly with volume mounts")
}
