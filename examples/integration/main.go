// Example showing integration with existing Gin application
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MK-Morse-SMS/Zoom-Alert"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize ZoomAlert module
	config := zoomalert.LoadConfigFromEnv()
	module, err := zoomalert.NewZoomAlertModule(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create Gin router
	router := gin.Default()

	// Your existing application routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "My Application with ZoomAlert Integration",
			"version": "1.0.0",
		})
	})

	// Business logic endpoint that sends Zoom alerts
	router.POST("/order", func(c *gin.Context) {
		var order struct {
			CustomerEmail string  `json:"customer_email"`
			Amount        float64 `json:"amount"`
			OrderID       string  `json:"order_id"`
		}

		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Process the order (your business logic here)
		// ...

		// Send notification to sales team via Zoom
		message := fmt.Sprintf("🛒 New Order Received!\nOrder ID: %s\nCustomer: %s\nAmount: $%.2f",
			order.OrderID, order.CustomerEmail, order.Amount)

		if err := module.SendAlert("sales@company.com", message); err != nil {
			log.Printf("Failed to send Zoom notification: %v", err)
			// Don't fail the order processing due to notification failure
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"order_id": order.OrderID,
		})
	})

	// Error reporting endpoint
	router.POST("/report-error", func(c *gin.Context) {
		var errorReport struct {
			Service   string `json:"service"`
			Error     string `json:"error"`
			Severity  string `json:"severity"`
			Timestamp string `json:"timestamp"`
		}

		if err := c.ShouldBindJSON(&errorReport); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Send error notification to development team
		var emoji string
		switch errorReport.Severity {
		case "critical":
			emoji = "🚨"
		case "high":
			emoji = "⚠️"
		default:
			emoji = "ℹ️"
		}

		message := fmt.Sprintf("%s Error Report\nService: %s\nSeverity: %s\nError: %s\nTime: %s",
			emoji, errorReport.Service, errorReport.Severity, errorReport.Error, errorReport.Timestamp)

		if err := module.SendAlert("kirwinr@mkmorse.com", message); err != nil {
			log.Printf("Failed to send error notification: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{"status": "reported"})
	})

	// Setup ZoomAlert routes (adds /api/v1/alert, /api/v1/health, etc.)
	module.RegisterOAuthRoutes(router)

	// Start server
	log.Printf("Starting server on port %s", config.Port)
	router.Run(":" + config.Port)
}
