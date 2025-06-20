package main

import (
	"log"
	"net/http"

	"github.com/kirwinrMK/Zoom-Alert"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config := zoomalert.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create module
	module, err := zoomalert.NewZoomAlertModule(config)
	if err != nil {
		log.Fatalf("Failed to create module: %v", err)
	}

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Register routes
	module.RegisterOAuthRoutes(router)
	module.RegisterAlertRoutes(router)

	// Add example endpoints documentation
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Zoom Alert Service",
			"version": "1.0.0",
			"endpoints": gin.H{
				"auth": gin.H{
					"GET /api/v1/auth/status":     "Check authorization status",
					"GET /api/v1/oauth/authorize": "Get OAuth authorization URL",
					"GET /api/v1/oauth/callback":  "OAuth callback handler",
				},
				"alerts": gin.H{
					"POST /api/v1/alert":           "Send simple text alert",
					"POST /api/v1/alert/rich":      "Send rich formatted alert",
					"POST /api/v1/alert/templated": "Send templated alert",
				},
				"health": gin.H{
					"GET /api/v1/health": "Health check",
				},
			},
			"examples": gin.H{
				"simple_alert": gin.H{
					"method": "POST",
					"url":    "/api/v1/alert",
					"body": gin.H{
						"email":   "user@example.com",
						"message": "This is a simple alert message",
					},
				},
				"rich_alert": gin.H{
					"method": "POST",
					"url":    "/api/v1/alert/rich",
					"body": gin.H{
						"email":        "user@example.com",
						"alert_text":   "System Alert: High CPU Usage Detected",
						"alert_level":  "ERROR",
						"section_text": "Server monitoring has detected high resource usage",
						"closeable":    true,
					},
				},
				"templated_alert": gin.H{
					"method": "POST",
					"url":    "/api/v1/alert/templated",
					"body": gin.H{
						"email":        "user@example.com",
						"alert_text":   "Database performance issue detected",
						"alert_level":  "WARNING",
						"section_text": "Query response times are above normal thresholds",
						"closeable":    true,
					},
				},
			},
		})
	})

	// Start server
	port := config.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Zoom Alert Server starting on port %s", port)
	log.Printf("📋 Visit http://localhost:%s for API documentation", port)
	log.Printf("🔐 OAuth Status: http://localhost:%s/api/v1/auth/status", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
