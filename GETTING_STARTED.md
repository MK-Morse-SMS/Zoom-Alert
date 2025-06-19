# Getting Started with ZoomAlert Module

This guide will help you integrate the ZoomAlert module into your Go project in just a few minutes.

## Prerequisites

- Go 1.21 or later
- A Zoom account with API access
- Basic familiarity with Go modules

## Step 1: Zoom App Setup

### Create a Zoom App

1. Go to [Zoom Marketplace](https://marketplace.zoom.us/develop/create)
2. Choose **Server-to-Server OAuth** (recommended) or **OAuth** for user authorization
3. Fill in your app details:
   - **App Name**: Your App Name
   - **Company Name**: Your Company
   - **Developer Email**: your@email.com

### Configure Scopes

Add these scopes to your app:
- `user:read` - To look up users by email
- `im_chat_message:write` - To send chat messages

### Get Your Credentials

From your Zoom app dashboard, copy:
- **Account ID**
- **Client ID**
- **Client Secret** 
- **Robot JID** (if using Server-to-Server OAuth)

## Step 2: Install the Module

### Option A: Use the module directly

```bash
# Add to your go.mod
go get github.com/yourusername/zoomalert
```

### Option B: Copy the module locally

If you prefer to include the module in your project:

```bash
# Copy the zoomalert directory to your project
cp -r /path/to/zoomalert ./zoomalert
```

Then in your go.mod:
```go
replace github.com/yourusername/zoomalert => ./zoomalert
require github.com/yourusername/zoomalert v0.0.0
```

## Step 3: Configure Environment Variables

Create a `.env` file or set environment variables:

```bash
# Required
export ZOOM_ACCOUNT_ID="your_account_id_here"
export ZOOM_CLIENT_ID="your_client_id_here"
export ZOOM_CLIENT_SECRET="your_client_secret_here"

# Optional
export ZOOM_ROBOT_JID="your_robot_jid_here"
export ZOOM_REDIRECT_URI="http://localhost:8080/api/v1/oauth/callback"
export PORT="8080"
export LOG_LEVEL="info"
```

## Step 4: Basic Usage

### Simplest Example

```go
package main

import (
    "log"
    "github.com/yourusername/zoomalert"
)

func main() {
    // Load config from environment
    config := zoomalert.LoadConfigFromEnv()
    
    // Initialize module
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Send an alert
    err = module.SendAlert("user@company.com", "Hello from ZoomAlert!")
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Println("Alert sent successfully!")
    }
}
```

### Complete Example with Error Handling

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/zoomalert"
)

func main() {
    // Load and validate configuration
    config := zoomalert.LoadConfigFromEnv()
    if err := config.Validate(); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    
    // Initialize module
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatalf("Failed to initialize ZoomAlert: %v", err)
    }
    
    // Check authorization status
    if !module.IsUserAuthorized() {
        fmt.Println("⚠️  User authorization required")
        authURL, err := module.GetAuthorizationURL()
        if err != nil {
            log.Fatalf("Failed to get auth URL: %v", err)
        }
        fmt.Printf("Please visit: %s\n", authURL)
        return
    }
    
    // Send alert
    email := "user@company.com"
    message := "🚀 Your application is now using ZoomAlert!"
    
    fmt.Printf("Sending alert to %s...\n", email)
    err = module.SendAlert(email, message)
    if err != nil {
        log.Printf("❌ Failed to send alert: %v", err)
    } else {
        fmt.Println("✅ Alert sent successfully!")
    }
}
```

## Step 5: Run Your Application

```bash
# Set environment variables
export ZOOM_ACCOUNT_ID="your_account_id"
export ZOOM_CLIENT_ID="your_client_id"
export ZOOM_CLIENT_SECRET="your_client_secret"

# Run your application
go run main.go
```

## Step 6: Handle OAuth Authorization

If you see "User authorization required", you need to complete the OAuth flow:

1. Run your application
2. Copy the authorization URL from the output
3. Visit the URL in your browser
4. Authorize the application
5. The module will handle the callback automatically

## Common Integration Patterns

### Pattern 1: HTTP Server

```go
package main

import (
    "context"
    "log"
    "github.com/yourusername/zoomalert"
)

func main() {
    config := zoomalert.LoadConfigFromEnv()
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start HTTP server with all endpoints
    ctx := context.Background()
    log.Println("Starting server on port", config.Port)
    log.Fatal(module.StartHTTPServer(ctx))
}
```

### Pattern 2: Add to Existing Gin Router

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/zoomalert"
)

func main() {
    // Your existing router
    router := gin.Default()
    
    // Your existing routes
    router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })
    
    // Add ZoomAlert routes
    config := zoomalert.LoadConfigFromEnv()
    module, _ := zoomalert.NewZoomAlertModule(config)
    module.SetupRoutes(router)
    
    router.Run(":8080")
}
```

### Pattern 3: Background Alert System

```go
package main

import (
    "log"
    "time"
    "github.com/yourusername/zoomalert"
)

type AlertSystem struct {
    zoomAlert *zoomalert.ZoomAlertModule
}

func NewAlertSystem() *AlertSystem {
    config := zoomalert.LoadConfigFromEnv()
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }
    
    return &AlertSystem{zoomAlert: module}
}

func (a *AlertSystem) MonitorSystem() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // Check system health
        if !a.isSystemHealthy() {
            a.zoomAlert.SendAlert("admin@company.com", "System unhealthy!")
        }
    }
}

func (a *AlertSystem) isSystemHealthy() bool {
    // Your health check logic
    return true
}

func main() {
    alertSystem := NewAlertSystem()
    alertSystem.MonitorSystem()
}
```

## Testing Your Integration

### Test Configuration

```go
func TestZoomAlertConfig(t *testing.T) {
    config := &zoomalert.Config{
        ZoomAccountID:    "test_account",
        ZoomClientID:     "test_client", 
        ZoomClientSecret: "test_secret",
    }
    
    err := config.Validate()
    if err != nil {
        t.Errorf("Config validation failed: %v", err)
    }
}
```

### Test Module Initialization

```go
func TestZoomAlertModule(t *testing.T) {
    config := &zoomalert.Config{
        ZoomAccountID:    "test_account",
        ZoomClientID:     "test_client",
        ZoomClientSecret: "test_secret",
    }
    
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        t.Fatalf("Failed to create module: %v", err)
    }
    
    if module == nil {
        t.Error("Module should not be nil")
    }
}
```

## Troubleshooting

### Common Issues

1. **"Configuration error: ZOOM_ACCOUNT_ID is required"**
   - Make sure all required environment variables are set
   - Check that your .env file is loaded correctly

2. **"User authorization required"**
   - Complete the OAuth flow by visiting the authorization URL
   - Ensure your redirect URI matches your Zoom app settings

3. **"User with email xxx not found"**
   - The email must belong to a user in your Zoom account
   - Check that the user exists and has the correct email

4. **"Failed to send chat message"**
   - Verify your Robot JID is correct
   - Check that your app has the required scopes

### Debug Mode

Enable debug logging to see more details:

```bash
export LOG_LEVEL="debug"
```

### Check Authorization Status

```bash
curl http://localhost:8080/api/v1/auth/status
```

## Next Steps

1. **Read the [README.md](README.md)** for complete API documentation
2. **Check [examples/](examples/)** for more integration patterns
3. **Review [MIGRATION.md](MIGRATION.md)** if upgrading from the standalone service
4. **File issues** on GitHub for any problems

## Quick Reference

### Environment Variables
```bash
ZOOM_ACCOUNT_ID=required
ZOOM_CLIENT_ID=required
ZOOM_CLIENT_SECRET=required
ZOOM_ROBOT_JID=optional
ZOOM_REDIRECT_URI=optional
PORT=optional
LOG_LEVEL=optional
```

### Main Functions
```go
// Initialize
config := zoomalert.LoadConfigFromEnv()
module, err := zoomalert.NewZoomAlertModule(config)

// Send alert
err = module.SendAlert(email, message)

// Check authorization
authorized := module.IsUserAuthorized()

// Get auth URL
authURL, err := module.GetAuthorizationURL()

// Start HTTP server
ctx := context.Background()
err = module.StartHTTPServer(ctx)
```

### HTTP Endpoints
- `GET /api/v1/health` - Health check
- `POST /api/v1/alert` - Send alert
- `GET /api/v1/auth/status` - Check auth status
- `GET /api/v1/oauth/authorize` - Get auth URL
- `GET /api/v1/oauth/callback` - OAuth callback

Happy coding! 🚀
