// Example application showing how to use the ZoomAlert module
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MK-Morse-SMS/Zoom-Alert"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 ZoomAlert Module Example")
	fmt.Println("==========================")

	// Load configuration from environment variables
	config := zoomalert.LoadConfigFromEnv()

	// Initialize the ZoomAlert module
	module, err := zoomalert.NewZoomAlertModule(config)
	if err != nil {
		log.Fatalf("Failed to initialize ZoomAlert module: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("\n🛑 Shutting down...")
		cancel()
	}()

	// Register OAuth routes
	fmt.Println("Press Ctrl+C to stop")
	router := gin.Default()
	module.RegisterOAuthRoutes(router)
	module.RegisterAlertRoutes(router)

	// Create and start the gin HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: router,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Check if user is already authorized
	fmt.Printf("User authorized: %v\n", module.IsUserAuthorized())

	// If not authorized, show how to get authorization URL
	if !module.IsUserAuthorized() {
		fmt.Println("\n🔐 OAuth Authorization Required")
		fmt.Println("To send alerts, you need to authorize the application.")

		authURL, err := module.GetAuthorizationURL()
		if err != nil {
			log.Fatalf("Failed to get authorization URL: %v", err)
		}

		fmt.Printf("Please visit this URL to authorize: %s\n", authURL)
		fmt.Println("Waiting for authorization...")

		// Wait until user is authorized
		for !module.IsUserAuthorized() {
			time.Sleep(2 * time.Second)
		}

		fmt.Println("✅ Authorization successful!")
	}

	// Example 1: Basic alert sending
	fmt.Println("\n📨 Example 1: Basic Alert Sending")
	err = module.SendAlert("kirwinr@mkmorse.com", "Hello from ZoomAlert module!")
	if err != nil {
		fmt.Printf("❌ Failed to send alert: %v\n", err)
	} else {
		fmt.Println("✅ Alert sent successfully!")
	}

	// // Example 2: Sending multiple alerts
	// fmt.Println("\n📨 Example 2: Multiple Alerts")
	// users := []string{"kirwinr@mkmorse.com", "whitmerl@mkmorse.com"}
	// for _, user := range users {
	// 	message := fmt.Sprintf("System notification: Server maintenance at %s", time.Now().Format("15:04"))
	// 	err := module.SendAlert(user, message)
	// 	if err != nil {
	// 		fmt.Printf("❌ Failed to send alert to %s: %v\n", user, err)
	// 	} else {
	// 		fmt.Printf("✅ Alert sent to %s\n", user)
	// 	}
	// }

	// Example 3: Using the HTTP server
	fmt.Println("\n🌐 Example 3: HTTP Server")
	fmt.Println("Starting HTTP server on port", config.Port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/v1/health")
	fmt.Println("  POST /api/v1/alert")
	fmt.Println("  GET  /api/v1/auth/status")
	fmt.Println("  GET  /api/v1/oauth/authorize")
	fmt.Println("  GET  /api/v1/oauth/callback")

	// Wait for shutdown signal
	<-ctx.Done()

	fmt.Println("👋 Goodbye!")
}
