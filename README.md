# ZoomAlert - Zoom Messaging Module for Go

ZoomAlert is a Go module that provides easy integration with Zoom's messaging platform. It allows you to send direct messages to Zoom users by email address through a simple API.

## Features

- 🚀 **Easy Integration**: Simple API for sending Zoom messages
- 🔐 **OAuth 2.0 Support**: Handles Zoom OAuth authentication automatically
- 🏗️ **Modular Design**: Can be integrated into existing projects or run standalone
- 📡 **HTTP API**: Optional HTTP server with REST endpoints
- 🛡️ **Security**: Built-in CSRF protection and token management
- 📝 **Logging**: Structured logging with configurable levels

## Installation

```bash
go get github.com/yourusername/zoomalert
```

## Quick Start

### Basic Usage (Programmatic)

```go
package main

import (
    "context"
    "log"
    "github.com/yourusername/zoomalert"
)

func main() {
    // Create configuration
    config := &zoomalert.Config{
        ZoomAccountID:    "your_account_id",
        ZoomClientID:     "your_client_id", 
        ZoomClientSecret: "your_client_secret",
        ZoomRobotJID:     "your_robot_jid",
        ZoomRedirectURI:  "http://localhost:8080/api/v1/oauth/callback",
    }

    // Initialize the module
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }

    // Send an alert (requires OAuth authorization first)
    err = module.SendAlert("user@company.com", "System alert: Server is down!")
    if err != nil {
        log.Printf("Failed to send alert: %v", err)
    }
}
```

### Using Environment Variables

```go
package main

import (
    "context"
    "log"
    "os"
    "github.com/yourusername/zoomalert"
)

func main() {
    // Load configuration from environment variables
    config := zoomalert.LoadConfigFromEnv()
    
    // Initialize the module
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }

    // Send an alert
    err = module.SendAlert("user@company.com", "Hello from ZoomAlert!")
    if err != nil {
        log.Printf("Failed to send alert: %v", err)
    }
}
```

### Standalone HTTP Server

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/yourusername/zoomalert"
)

func main() {
    config := zoomalert.LoadConfigFromEnv()
    
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }

    // Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown signals
    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        cancel()
    }()

    // Start HTTP server
    log.Println("Starting ZoomAlert HTTP server...")
    if err := module.StartHTTPServer(ctx); err != nil {
        log.Printf("Server error: %v", err)
    }
}
```

### Integration with Existing Gin Router

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/zoomalert"
)

func main() {
    config := zoomalert.LoadConfigFromEnv()
    
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }

    // Create your existing Gin router
    router := gin.Default()
    
    // Add your existing routes
    router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })

    // Setup ZoomAlert routes
    module.SetupRoutes(router)

    // Start server
    router.Run(":8080")
}
```

## Configuration

The module can be configured using environment variables or programmatically:

### Environment Variables

```bash
export ZOOM_ACCOUNT_ID="your_zoom_account_id"
export ZOOM_CLIENT_ID="your_zoom_client_id"  
export ZOOM_CLIENT_SECRET="your_zoom_client_secret"
export ZOOM_ROBOT_JID="your_zoom_robot_jid"
export ZOOM_REDIRECT_URI="http://localhost:8080/api/v1/oauth/callback"
export PORT="8080"
export LOG_LEVEL="info"  # debug, info, warn, error
```

### Programmatic Configuration

```go
config := &zoomalert.Config{
    ZoomAccountID:    "your_account_id",
    ZoomClientID:     "your_client_id",
    ZoomClientSecret: "your_client_secret", 
    ZoomRobotJID:     "your_robot_jid",
    ZoomRedirectURI:  "http://localhost:8080/api/v1/oauth/callback",
    Port:             "8080",
    LogLevel:         "info",
}
```

## HTTP API Endpoints

When using the HTTP server, the following endpoints are available:

| Method | Endpoint                    | Description                          |
|--------|-----------------------------|--------------------------------------|
| GET    | `/api/v1/health`           | Health check                         |
| POST   | `/api/v1/alert`            | Send alert message                   |
| GET    | `/api/v1/auth/status`      | Check authorization status           |
| GET    | `/api/v1/oauth/authorize`  | Get OAuth authorization URL          |
| GET    | `/api/v1/oauth/callback`   | OAuth callback handler               |

### Send Alert

```bash
curl -X POST http://localhost:8080/api/v1/alert \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "message": "System alert: High CPU usage detected!"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Alert sent successfully"
}
```

### Check Authorization Status

```bash
curl http://localhost:8080/api/v1/auth/status
```

**Response:**
```json
{
  "user_authorized": true,
  "message": "User authorization available - full user lookup enabled"
}
```

### Get OAuth Authorization URL

```bash
curl http://localhost:8080/api/v1/oauth/authorize
```

**Response:**
```json
{
  "authorization_url": "https://zoom.us/oauth/authorize?...",
  "state": "secure_random_state",
  "message": "Please visit the authorization URL to complete OAuth flow"
}
```

## Zoom Setup

1. **Create a Zoom App**:
   - Go to [Zoom Marketplace](https://marketplace.zoom.us/)
   - Create a new **Server-to-Server OAuth** app (for backend use) or **OAuth** app (for user authorization)

2. **Configure Scopes**:
   - `user:read` - To look up users by email
   - `im_chat_message:write` - To send chat messages

3. **Get Credentials**:
   - Account ID
   - Client ID  
   - Client Secret
   - Robot JID (for chatbot apps)

4. **Set Redirect URI**:
   - Add your callback URL: `http://localhost:8080/api/v1/oauth/callback`

## OAuth Flow

The module handles OAuth authentication automatically:

1. Call `/api/v1/oauth/authorize` to get the authorization URL
2. User visits the URL and authorizes the app
3. Zoom redirects to your callback URL with the authorization code
4. Module exchanges the code for access tokens
5. Tokens are stored and refreshed automatically

## Advanced Usage

### Custom Logging

```go
import "log/slog"

module, _ := zoomalert.NewZoomAlertModule(config)

// Access the module's logger
logger := module.Logger()
logger.Info("Custom log message", "key", "value")
```

### Direct Service Access

```go
module, _ := zoomalert.NewZoomAlertModule(config)

// Access underlying services for advanced usage
zoomService := module.GetZoomService()
oauthService := module.GetOAuthService()

// Use services directly
user, err := zoomService.GetUserByEmail("user@company.com")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User: %s (%s)\n", user.FirstName, user.Email)
```

### Integration Examples

#### Monitoring System Integration

```go
// Prometheus alertmanager webhook
func alertHandler(w http.ResponseWriter, r *http.Request) {
    var alert PrometheusAlert
    json.NewDecoder(r.Body).Decode(&alert)
    
    for _, alert := range alert.Alerts {
        message := fmt.Sprintf("🚨 Alert: %s\nStatus: %s\nDescription: %s", 
            alert.Labels["alertname"], 
            alert.Status,
            alert.Annotations["description"])
            
        module.SendAlert("devops@company.com", message)
    }
}
```

#### CI/CD Pipeline Integration

```go
// GitHub Actions / Jenkins integration
func deploymentNotification(deploymentStatus string, repoName string) {
    var message string
    if deploymentStatus == "success" {
        message = fmt.Sprintf("✅ Deployment successful for %s", repoName)
    } else {
        message = fmt.Sprintf("❌ Deployment failed for %s", repoName)
    }
    
    module.SendAlert("team@company.com", message)
}
```

## Error Handling

The module provides detailed error information:

```go
err := module.SendAlert("invalid-email", "test message")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "not found"):
        log.Println("User not found in Zoom")
    case strings.Contains(err.Error(), "authorization required"):
        log.Println("Need to complete OAuth flow first")
    default:
        log.Printf("Other error: %v", err)
    }
}
```

## Testing

```bash
# Run tests
go test ./...

# Test with coverage
go test -cover ./...

# Integration test (requires valid Zoom credentials)
go test -tags=integration ./...
```

## Development

### Building from Source

```bash
git clone https://github.com/yourusername/zoomalert
cd zoomalert
go mod tidy
go build
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- 📖 [Zoom API Documentation](https://developers.zoom.us/docs/)
- 🐛 [Report Issues](https://github.com/yourusername/zoomalert/issues)
- 💬 [Discussions](https://github.com/yourusername/zoomalert/discussions)

---

**Note**: Make sure to keep your Zoom credentials secure and never commit them to version control. Use environment variables or secure secret management systems in production.
