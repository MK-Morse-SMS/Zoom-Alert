# ZoomAlert - Zoom Messaging Module for Go

ZoomAlert is a Go module that provides easy integration with Zoom's messaging platform. It allows you to send direct messages and rich formatted alerts to Zoom users by email address through both programmatic API and HTTP endpoints.

## Features

- 🚀 **Easy Integration**: Simple API for sending Zoom messages
- 🔐 **OAuth 2.0 Support**: Handles Zoom OAuth authentication automatically  
- 🏗️ **Modular Design**: Can be integrated into existing projects or run standalone
- 📡 **HTTP API**: REST endpoints for alert operations
- 🎨 **Rich Alerts**: Support for formatted alerts with different severity levels
- 🛡️ **Security**: Built-in CSRF protection and token management
- 📝 **Logging**: Structured logging with configurable levels

## Getting Started

### Prerequisites

- Go 1.21 or later
- A Zoom account with API access
- Basic familiarity with Go modules

### Installation

This module is designed to be used locally in your project:

```bash
# Copy the zoomalert module to your project
cp -r /path/to/zoomalert ./
```

Add to your `go.mod`:

```go
module your-project

require zoomalert v0.0.0
replace zoomalert => ./zoomalert
```

### Environment Setup

Set up your environment variables:

```bash
# Required
export ZOOM_ACCOUNT_ID="your_account_id_here"
export ZOOM_CLIENT_ID="your_client_id_here"
export ZOOM_CLIENT_SECRET="your_client_secret_here"
export ZOOM_ROBOT_JID="your_robot_jid_here"

# Optional
export ZOOM_REDIRECT_URL="http://localhost:8080/api/v1/oauth/callback"
export PORT="8080"
export LOG_LEVEL="info"
```

### Basic Usage

#### Programmatic API

```go
package main

import (
    "log"
    "zoomalert"
)

func main() {
    // Load config from environment
    config := zoomalert.LoadConfigFromEnv()
    
    // Initialize module
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Send a simple alert
    err = module.SendAlert("user@company.com", "Hello from ZoomAlert!")
    if err != nil {
        log.Printf("Failed to send alert: %v", err)
    }
}
```

#### HTTP Server

```go
package main

import (
    "context"
    "log"
    "zoomalert"
)

func main() {
    config := zoomalert.LoadConfigFromEnv()
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    log.Println("Starting ZoomAlert HTTP server...")
    if err := module.StartHTTPServer(ctx); err != nil {
        log.Printf("Server error: %v", err)
    }
}
```

#### Integration with Existing Gin Router

```go
package main

import (
    "github.com/gin-gonic/gin"
    "zoomalert"
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

## Zoom App Setup

### 1. Create a Zoom App

1. Go to [Zoom Marketplace](https://marketplace.zoom.us/develop/create)
2. Choose **Server-to-Server OAuth** (recommended) or **OAuth** for user authorization
3. Fill in your app details:
   - **App Name**: Your App Name
   - **Company Name**: Your Company
   - **Developer Email**: `your@email.com`

### 2. Configure Scopes

Add these required scopes to your app:

- `user:read` - To look up users by email
- `im_chat_message:write` - To send chat messages

### 3. Get Your Credentials

From your Zoom app dashboard, copy:

- **Account ID**
- **Client ID**
- **Client Secret**
- **Robot JID** (if using Server-to-Server OAuth)

### 4. Set Redirect URI

Add your callback URL: `http://localhost:8080/api/v1/oauth/callback`

## OAuth Flow

The module handles OAuth authentication automatically:

1. Call `/api/v1/oauth/authorize` to get the authorization URL
2. User visits the URL and authorizes the app  
3. Zoom redirects to your callback URL with the authorization code
4. Module exchanges the code for access tokens
5. Tokens are stored and refreshed automatically

## API Endpoints

When using the HTTP server, the following endpoints are available:

| Method | Endpoint                    | Description                          |
|--------|-----------------------------|--------------------------------------|
| GET    | `/api/v1/health`           | Health check                         |
| POST   | `/api/v1/alert`            | Send simple text alert               |
| POST   | `/api/v1/alert/rich`       | Send rich formatted alert            |
| POST   | `/api/v1/alert/templated`  | Send templated alert                 |
| GET    | `/api/v1/auth/status`      | Check authorization status           |
| GET    | `/api/v1/oauth/authorize`  | Get OAuth authorization URL          |
| GET    | `/api/v1/oauth/callback`   | OAuth callback handler               |

### API Examples

#### Send Simple Alert

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

#### Send Rich Formatted Alert

```bash
curl -X POST http://localhost:8080/api/v1/alert/rich \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "alert_text": "System Alert: High CPU Usage",
    "alert_level": "ERROR",
    "section_text": "Server monitoring detected high usage",
    "closeable": true
  }'
```

#### Check Authorization Status

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

#### Get OAuth Authorization URL

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

## Rich Formatted Alerts

ZoomAlert supports rich formatted alerts with different severity levels and structured content.

### Alert Levels

- **INFO**: General information and notifications
- **WARNING**: Warning conditions requiring attention
- **ERROR**: Error conditions that need to be addressed  
- **CRITICAL**: Critical alerts requiring immediate action

### Programmatic Usage

```go
// Send a rich formatted alert
err := zoomService.SendAlertWithRichContent(
    "user@company.com",
    "System Alert: High CPU Usage Detected",
    "ERROR", 
    true, // closeable
    "Server monitoring has detected high resource usage",
)

// Using templates for consistent formatting
alertContent := zoomalert.CreateAlertTemplate(
    "Database performance monitoring",
    "Warning: Query response time above threshold", 
    zoomalert.AlertLevelWarning,
    true,
)
err := zoomService.SendTemplatedAlert(userJID, alertContent)
```

### Alert Structure

The alert system supports rich content messages with the following structure:

```json
{
  "robot_jid": "your_robot_jid_here",
  "to_jid": "user_jid_here", 
  "account_id": "your_account_id_here",
  "content": {
    "settings": {},
    "body": [
      {
        "type": "section",
        "layout": "horizontal",
        "sections": [
          {
            "type": "message",
            "text": "Context message for the alert"
          }
        ]
      },
      {
        "type": "alert",
        "text": "Alert message text",
        "level": "ERROR|WARNING|INFO|CRITICAL",
        "closeable": true
      }
    ]
  }
}
```

## Module Configuration

### Environment Variables

```bash
# Required
ZOOM_ACCOUNT_ID="your_zoom_account_id"
ZOOM_CLIENT_ID="your_zoom_client_id"  
ZOOM_CLIENT_SECRET="your_zoom_client_secret"
ZOOM_ROBOT_JID="your_zoom_robot_jid"

# Optional
ZOOM_REDIRECT_URL="http://localhost:8080/api/v1/oauth/callback"
TARGET_EMAIL="default@company.com"
PORT="8080"
LOG_LEVEL="info"  # debug, info, warn, error
```

### Programmatic Setup

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

## Examples

The module includes several example applications in the `examples/` directory:

### Basic Example

```bash
cd examples/basic
go run main.go
```

Shows basic integration with HTTP server and programmatic usage.

### Alert Example  

```bash
cd examples/alert
go run main.go
```

Demonstrates sending rich formatted alerts.

### Integration Example

```bash
cd examples/integration
go run main.go
```

Shows integration with monitoring systems like Prometheus.

### Web Server Example

```bash
cd examples/web-server
go run main.go
```

Complete web application with frontend and API.

## Testing

### Test Scripts

Run the included test scripts to verify the endpoints:

```bash
# PowerShell (Windows)
.\test-api.ps1

# Bash (Linux/macOS)  
./test-api.sh
```

### Unit Tests

```bash
# Run tests
go test ./...

# Test with coverage
go test -cover ./...

# Integration test (requires valid Zoom credentials)
go test -tags=integration ./...
```

## Advanced Usage

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

### Custom Logging

```go
import "log/slog"

module, _ := zoomalert.NewZoomAlertModule(config)

// Access the module's logger
logger := module.Logger()
logger.Info("Custom log message", "key", "value")
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

## Development

### Building from Source

```bash
git clone <repository-url>
cd zoomalert
go mod tidy
go build
```

### Project Structure

```text
zoomalert/
├── README.md              # This file
├── go.mod                 # Go module definition
├── module.go              # Main module implementation
├── oauth.go               # OAuth service
├── zoom.go                # Zoom API service  
├── handlers.go            # HTTP handlers
├── cmd/
│   └── cli/
│       └── main.go        # CLI application
└── examples/
    ├── alert/             # Alert example
    ├── basic/             # Basic usage example
    ├── integration/       # Integration example
    └── web-server/        # Web server example
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
- 🐛 Report Issues: Create an issue in the repository
- 💬 Questions: Use GitHub Discussions

---

**Security Note**: Always keep your Zoom credentials secure and never commit them to version control. Use environment variables or secure secret management systems in production.
