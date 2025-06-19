# Migration Guide: From Standalone Service to Module

This guide helps you migrate from the original standalone Zoom Alert Service to the new modular version.

## Overview of Changes

The original project has been refactored into a reusable Go module with the following benefits:

- **Modular Design**: Can be integrated into existing applications
- **Programmatic API**: Direct function calls without HTTP overhead
- **Flexible Configuration**: Environment variables or programmatic setup
- **Better Testing**: Easier to unit test and mock
- **Cleaner Architecture**: Separated concerns and better organization

## Quick Migration

### Option 1: Drop-in Replacement (Minimal Changes)

If you want to keep using the HTTP API with minimal changes:

**Before (main.go):**
```go
// Old standalone service
func main() {
    cfg := config.Load()
    oauthService := oauth.NewOAuthService(cfg)
    zoomService := zoom.NewZoomService(oauthService, cfg.ZoomRobotJID, cfg.ZoomAccountID)
    alertHandler := handlers.NewAlertHandler(zoomService)
    
    router := gin.Default()
    // ... setup routes
    router.Run(":" + cfg.Port)
}
```

**After (main.go):**
```go
// New module-based service
import "github.com/yourusername/zoomalert"

func main() {
    config := zoomalert.LoadConfigFromEnv()
    module, err := zoomalert.NewZoomAlertModule(config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    module.StartHTTPServer(ctx)
}
```

### Option 2: Integration into Existing Application

**Before (separate service):**
```bash
# Run as separate service
go run main.go

# Call from your app
curl -X POST http://zoom-service:8080/alert \
  -d '{"email":"user@example.com","message":"Alert"}'
```

**After (integrated):**
```go
// In your existing application
import "github.com/yourusername/zoomalert"

func main() {
    // Your existing app setup
    app := gin.Default()
    
    // Initialize ZoomAlert module
    config := zoomalert.LoadConfigFromEnv()
    zoomModule, _ := zoomalert.NewZoomAlertModule(config)
    
    // Add ZoomAlert routes to your app
    zoomModule.SetupRoutes(app)
    
    // Your existing routes
    app.GET("/", myHandler)
    
    app.Run(":8080")
}
```

### Option 3: Programmatic Usage

**Before (HTTP calls):**
```go
// Old way: HTTP API calls
resp, err := http.Post("http://zoom-service:8080/alert", 
    "application/json",
    strings.NewReader(`{"email":"user@example.com","message":"Alert"}`))
```

**After (direct calls):**
```go
// New way: Direct function calls
import "github.com/yourusername/zoomalert"

config := zoomalert.LoadConfigFromEnv()
module, _ := zoomalert.NewZoomAlertModule(config)

err := module.SendAlert("user@example.com", "Alert")
```

## Configuration Migration

### Environment Variables

The environment variables remain the same:

```bash
# No changes needed
ZOOM_ACCOUNT_ID=your_account_id
ZOOM_CLIENT_ID=your_client_id
ZOOM_CLIENT_SECRET=your_client_secret
ZOOM_ROBOT_JID=your_robot_jid
ZOOM_REDIRECT_URI=http://localhost:8080/api/v1/oauth/callback
PORT=8080
```

### New Optional Variables

```bash
# New optional variables
LOG_LEVEL=info  # debug, info, warn, error
```

## API Compatibility

### HTTP Endpoints

All HTTP endpoints remain the same:

| Endpoint | Method | Status |
|----------|---------|---------|
| `/api/v1/health` | GET | ✅ Compatible |
| `/api/v1/alert` | POST | ✅ Compatible |
| `/api/v1/auth/status` | GET | ✅ Compatible |
| `/api/v1/oauth/authorize` | GET | ✅ Compatible |
| `/api/v1/oauth/callback` | GET | ✅ Compatible |

### Request/Response Formats

All request and response formats remain unchanged:

```json
// POST /api/v1/alert
{
  "email": "user@example.com",
  "message": "Your message here"
}

// Response
{
  "success": true,
  "message": "Alert sent successfully"
}
```

## Docker Migration

### Before (Dockerfile)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o zoom-alert-service .

FROM alpine:latest
COPY --from=builder /app/zoom-alert-service .
CMD ["./zoom-alert-service"]
```

### After (Dockerfile)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o zoom-alert-service ./examples/basic

FROM alpine:latest
COPY --from=builder /app/zoom-alert-service .
CMD ["./zoom-alert-service"]
```

Or create a custom main.go using the module:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy your custom main.go
COPY main.go .
RUN go build -o app .

FROM alpine:latest
COPY --from=builder /app/app .
CMD ["./app"]
```

## Testing Migration

### Before (integration tests)

```bash
# Start service
go run main.go &

# Test HTTP endpoints
curl http://localhost:8080/health
curl -X POST http://localhost:8080/alert -d '{"email":"test@example.com","message":"test"}'
```

### After (unit + integration tests)

```go
// Unit testing
func TestSendAlert(t *testing.T) {
    config := &zoomalert.Config{
        ZoomAccountID:    "test_account",
        ZoomClientID:     "test_client",
        ZoomClientSecret: "test_secret",
    }
    
    module, err := zoomalert.NewZoomAlertModule(config)
    assert.NoError(t, err)
    
    // Test logic without HTTP overhead
    err = module.SendAlert("test@example.com", "test message")
    // ... assertions
}
```

## Directory Structure Migration

### Before

```
zoom-alert-service/
├── main.go
├── config/
│   └── config.go
├── handlers/
│   └── handlers.go
├── oauth/
│   └── oauth.go
├── zoom/
│   └── zoom.go
└── templates/
```

### After

```
your-project/
├── main.go                    # Your application
├── go.mod                     # Add zoomalert dependency
└── vendor/                    # Or use go modules
    └── github.com/yourusername/zoomalert/
        ├── module.go
        ├── oauth.go
        ├── zoom.go
        ├── handlers.go
        └── examples/
```

## Troubleshooting

### Issue: Module not found

```bash
go: github.com/yourusername/zoomalert@latest: module not found
```

**Solution:** Update the import path to match your actual module path, or use local development:

```go
// Use replace directive for local development
// go.mod
replace github.com/yourusername/zoomalert => ./zoomalert

require github.com/yourusername/zoomalert v0.0.0
```

### Issue: Configuration not loading

```bash
Configuration error: ZOOM_ACCOUNT_ID is required
```

**Solution:** Ensure environment variables are set. The module uses the same environment variables as before.

### Issue: OAuth callback not working

**Solution:** Make sure your redirect URI in Zoom app settings matches your callback URL. The module uses the same OAuth flow.

## Performance Improvements

### Before (HTTP overhead)

```
Your App → HTTP → Zoom Service → HTTP → Zoom API
         ↑ Network latency     ↑ Network latency
```

### After (direct calls)

```
Your App → Module → Zoom API
         ↑ Function call (faster)
```

### Benefits

- **Reduced Latency**: No HTTP overhead for internal calls
- **Better Error Handling**: Direct error returns without HTTP status codes
- **Type Safety**: Compile-time checking instead of runtime HTTP errors
- **Easier Testing**: Mock dependencies at the interface level

## Advanced Migration Patterns

### Pattern 1: Gradual Migration

Keep the old HTTP service for existing clients while adding direct integration:

```go
func main() {
    config := zoomalert.LoadConfigFromEnv()
    module, _ := zoomalert.NewZoomAlertModule(config)
    
    // Direct usage for new features
    go func() {
        for alert := range alertChannel {
            module.SendAlert(alert.Email, alert.Message)
        }
    }()
    
    // Keep HTTP API for backward compatibility
    ctx := context.Background()
    module.StartHTTPServer(ctx)
}
```

### Pattern 2: Custom Wrapper

Create a wrapper that matches your existing interface:

```go
type LegacyAlertService struct {
    module *zoomalert.ZoomAlertModule
}

func (s *LegacyAlertService) SendZoomMessage(email, message string) error {
    return s.module.SendAlert(email, message)
}

// Use existing interface
var alertService AlertServiceInterface = &LegacyAlertService{module: zoomModule}
```

### Pattern 3: Microservice to Library

Transform from microservice architecture to library:

```go
// Before: Multiple services communicating via HTTP
// Service A → HTTP → Zoom Service → HTTP → Zoom API
// Service B → HTTP → Zoom Service → HTTP → Zoom API

// After: Shared library
// Service A → ZoomAlert Module → Zoom API
// Service B → ZoomAlert Module → Zoom API

// Each service integrates the module directly
func initializeServices() {
    zoomConfig := zoomalert.LoadConfigFromEnv()
    zoomModule, _ := zoomalert.NewZoomAlertModule(zoomConfig)
    
    serviceA := NewServiceA(zoomModule)
    serviceB := NewServiceB(zoomModule)
    
    // Each service can send alerts directly
}
```

## Rollback Plan

If you need to rollback to the original standalone service:

1. Keep the original code in a separate branch
2. The module can be used to create a standalone service identical to the original
3. Environment variables and API remain the same for easy rollback

## Need Help?

- Check the [README.md](README.md) for detailed documentation
- Look at [examples/](examples/) for common patterns
- File an issue for migration-specific problems
