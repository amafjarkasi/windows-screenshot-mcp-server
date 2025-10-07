# Screenshot MCP Server - Examples

This directory contains practical examples and tools for using the WebSocket streaming functionality of the Screenshot MCP Server.

## 📁 Contents

### 🌐 HTML WebSocket Viewer (`websocket-viewer.html`)
An interactive web-based viewer for real-time screenshot streaming.

**Features:**
- Live WebSocket streaming with configurable FPS and quality
- Real-time statistics (frame count, data usage, connection status)
- Dynamic settings adjustment (FPS, quality, image format)
- Responsive design for mobile and desktop
- Connection status indicators and logging

**Usage:**
1. Start the screenshot server: `go run cmd/server/main.go`
2. Open `websocket-viewer.html` in your browser
3. Enter a valid window ID and click "Start Stream"
4. Adjust settings in real-time using the controls

**URL Parameters:**
- `?server=localhost:8080` - Set server URL
- `?windowId=123456` - Pre-populate window ID

### 🔧 Go Streaming Client (`streaming-client/`)
A programmatic Go client for WebSocket streaming integration.

**Features:**
- Command-line interface for streaming
- Structured logging with emoji indicators
- Real-time option updates (FPS, quality, format)
- Graceful shutdown handling
- Context-aware timeouts

**Usage:**
```bash
cd streaming-client
go mod tidy
go run main.go -windows 123456 -server localhost:8080 -timeout 30s
```

**Options:**
- `-server` - Server URL (default: localhost:8080)
- `-windows` - Comma-separated window IDs
- `-timeout` - Connection timeout (default: 30s)
- `-verbose` - Enable verbose logging

### 🧪 PowerShell Test Script (`test-streaming.ps1`)
Comprehensive testing script for WebSocket streaming functionality.

**Features:**
- Server connection validation
- Window enumeration and validation
- Automated WebSocket testing
- Browser viewer launching
- Go client execution

**Usage:**
```powershell
# Basic test
.\test-streaming.ps1 -WindowId 123456

# Full test with browser and Go client
.\test-streaming.ps1 -WindowId 123456 -OpenBrowser -RunGoClient -ShowLogs

# Custom server and duration
.\test-streaming.ps1 -WindowId 123456 -ServerUrl "localhost:9000" -TestDuration 60
```

**Parameters:**
- `-WindowId` - Target window ID (required)
- `-ServerUrl` - Server URL (default: localhost:8080)
- `-TestDuration` - Test duration in seconds (default: 30)
- `-OpenBrowser` - Launch HTML viewer in browser
- `-RunGoClient` - Run Go streaming client
- `-ShowLogs` - Enable verbose logging

## 🚀 Getting Started

### 1. Start the Server
```bash
# From project root
go run cmd/server/main.go
```

The server will start on `localhost:8080` by default.

### 2. Get Available Windows
```bash
# List all windows
curl http://localhost:8080/api/windows

# Or using the MCP API
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "window.list", "id": 1}'
```

### 3. Start Streaming

#### Option A: HTML Viewer
1. Open `websocket-viewer.html` in your browser
2. Enter a window ID from the list
3. Click "Start Stream"

#### Option B: Go Client
```bash
cd streaming-client
go run main.go -windows YOUR_WINDOW_ID
```

#### Option C: Direct WebSocket Connection
Connect to: `ws://localhost:8080/stream/{windowId}`

## 🔧 API Endpoints

### HTTP REST API
- `GET /api/health` - Server health check
- `GET /api/windows` - List available windows
- `GET /api/screenshot?window=ID&format=png` - Take single screenshot

### WebSocket Streaming API
- `WS /stream/{windowId}` - Start streaming session
- `GET /v1/stream/status` - Get streaming statistics

### MCP JSON-RPC API
- `POST /rpc` - MCP 2.0 JSON-RPC endpoint
  - `window.list` - List windows
  - `screenshot.capture` - Take screenshot
  - `stream.status` - Get streaming status

## 🎛️ Configuration

### Server Configuration
The server can be configured through environment variables or config file:

```json
{
  "port": 8080,
  "host": "localhost",
  "default_format": "png",
  "quality": 95,
  "stream_max_sessions": 10,
  "stream_default_fps": 10
}
```

### WebSocket Query Parameters
When connecting to the WebSocket endpoint, you can specify:
- `fps` - Frames per second (1-60)
- `quality` - Image quality (1-100)
- `format` - Image format (png, jpeg, webp)

Example: `ws://localhost:8080/stream/123456?fps=15&quality=90&format=jpeg`

## 🐛 Troubleshooting

### Common Issues

**Connection Refused**
- Ensure the server is running on the correct port
- Check Windows Firewall settings
- Verify the window ID exists

**Poor Streaming Performance**
- Reduce FPS or image quality
- Use JPEG format for better compression
- Close unnecessary applications

**Window Not Found**
- Use the `/api/windows` endpoint to get valid window IDs
- Ensure the target window is visible and not minimized

### Testing WebSocket Connection

**Using wscat (if available):**
```bash
npm install -g wscat
wscat -c ws://localhost:8080/stream/123456
```

**Using PowerShell:**
```powershell
# Test server connectivity
Test-NetConnection localhost -Port 8080

# Get windows list
Invoke-RestMethod http://localhost:8080/api/windows
```

## 📊 Performance Tips

### For High Frame Rates
- Use JPEG format with 70-80% quality
- Limit to essential windows only
- Monitor CPU and memory usage

### For High Quality
- Use PNG format for screenshots with text
- Reduce FPS to 5-10 for detailed viewing
- Consider the network bandwidth

### For Multiple Streams
- Monitor the `stream.status` endpoint
- Implement client-side connection pooling
- Use different quality settings per stream

## 🔗 Integration Examples

### Automation Scripts
```bash
# Take screenshot and start streaming
WINDOW_ID=$(curl -s http://localhost:8080/api/windows | jq -r '.windows[0].ID')
curl "http://localhost:8080/api/screenshot?window=$WINDOW_ID&format=png" -o screenshot.png
```

### Monitoring Dashboards
The WebSocket streaming can be integrated into monitoring tools, remote access applications, or automated testing frameworks.

### Development Workflows
Use the streaming API for:
- Visual regression testing
- UI development feedback loops
- Remote debugging sessions
- Live demos and presentations

## 📝 Notes

- Windows 10/11 required for full functionality
- Administrative privileges may be needed for some windows
- Large windows or high frame rates will consume more bandwidth
- WebSocket connections are limited by server configuration

For more information, see the main project documentation and API reference.