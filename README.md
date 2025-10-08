# Windows Screenshot MCP Server

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP Compatible](https://img.shields.io/badge/MCP-Compatible-green.svg)](https://github.com/modelcontextprotocol/specification)

**Professional Windows screenshot capture server with Model Context Protocol (MCP) integration, real-time WebSocket streaming, Chrome tab capture, and advanced window targeting.**

## Overview

A production-ready Go-based screenshot server that provides both REST API and MCP protocol support for capturing Windows application screenshots. Designed for automation, testing, and AI agent integration with advanced features like real-time streaming and hidden window capture.

## Features

### Core Screenshot Capabilities
- **Window targeting**: Capture by title, class name, process ID, or window handle
- **Multiple image formats**: PNG, JPEG, BMP, WebP with configurable quality
- **Region capture**: Specify rectangular areas for precise screenshots
- **Advanced window handling**: Support for hidden, minimized, and system tray applications

### Chrome Browser Integration
- **Tab discovery**: Automatically find Chrome instances and enumerate tabs
- **Direct tab capture**: Screenshot specific browser tabs via Chrome DevTools
- **Multiple Chrome support**: Handle multiple Chrome processes simultaneously

### Real-Time WebSocket Streaming
- **Live streaming**: Real-time window feeds via WebSocket connections
- **Configurable quality**: Adjust FPS (1-60), quality, and format dynamically
- **Multiple sessions**: Support concurrent streaming sessions
- **Session management**: Start, stop, and monitor active streaming sessions

### Dual Protocol Support
- **REST API**: Traditional HTTP endpoints for easy integration
- **Model Context Protocol (MCP)**: JSON-RPC 2.0 for AI agent integration
- **Health monitoring**: Built-in health checks and status reporting
- **CORS support**: Cross-origin requests enabled for web applications

## Quick Start

### Installation

```bash
# Download latest release
curl -L https://github.com/your-org/screenshot-mcp-server/releases/latest/download/screenshot-server.exe -o screenshot-server.exe

# Or build from source
git clone https://github.com/your-org/screenshot-mcp-server.git
cd screenshot-mcp-server
go build -o screenshot-server.exe ./cmd/server
```

### Basic Usage

```bash
# Start the server
./screenshot-server.exe --port 8080

# Health check
curl http://localhost:8080/health

# Basic window capture
curl "http://localhost:8080/api/screenshot?method=title&target=Notepad" -o notepad.png

# Full desktop capture
curl "http://localhost:8080/api/screenshot?method=desktop&monitor=0" -o desktop.png
```

## API Reference

### REST Endpoints

#### Health Check
```http
GET /health
```
Returns server status and version information.

#### Screenshot Capture
```http
GET /api/screenshot
GET /v1/screenshot
```

**Parameters:**
- `method` (required): `title`, `pid`, `handle`, `class`
- `target` (required): Window identifier (title, PID, handle, class name)
- `format`: `png`, `jpeg`, `bmp`, `webp` (default: `png`)
- `quality`: 1-100 for lossy formats (default: 95)
- `cursor`: `true`/`false` to include mouse cursor

**Examples:**
```bash
# Window by title
curl "http://localhost:8080/api/screenshot?method=title&target=Calculator" -o calc.png

# Window by PID
curl "http://localhost:8080/api/screenshot?method=pid&target=1234&format=jpeg&quality=80" -o app.jpg

# Window by class name
curl "http://localhost:8080/api/screenshot?method=class&target=Notepad&cursor=true" -o notepad.png
```

#### Chrome Integration
```http
GET /v1/chrome/instances          # List Chrome instances
GET /v1/chrome/tabs               # List all Chrome tabs
POST /v1/chrome/tabs/:id/screenshot  # Capture specific tab
```

### WebSocket Streaming

Connect to `ws://localhost:8080/stream/{windowId}` for real-time streaming.

**Query Parameters:**
- `fps`: Frames per second (1-60, default: 10)
- `quality`: Compression quality (10-100, default: 75)
- `format`: `jpeg` or `png` (default: `jpeg`)

**Client Example:**
```html
<!DOCTYPE html>
<html>
<body>
    <img id="stream" style="max-width: 100%;">
    <script>
        const ws = new WebSocket('ws://localhost:8080/stream/0?fps=15&quality=75&format=jpeg');
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'frame') {
                document.getElementById('stream').src = 'data:image/jpeg;base64,' + data.image;
            }
        };
    </script>
</body>
</html>
```

### Model Context Protocol (MCP)

The server supports MCP JSON-RPC 2.0 requests via `POST /rpc`.

**Available Methods:**
- `screenshot.capture` - Capture screenshots
- `window.list` - List windows (placeholder)
- `chrome.instances` - List Chrome instances
- `chrome.tabs` - List Chrome tabs
- `chrome.tabCapture` - Capture Chrome tab
- `stream.status` - Get streaming status

**Example MCP Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "screenshot.capture",
  "params": {
    "method": "title",
    "target": "Calculator",
    "format": "png"
  },
  "id": 1
}
```

### Server Configuration

The server can be configured via environment variables or command-line flags:

```bash
# Start with custom port
./server.exe --port 9090

# Start with custom host
./server.exe --host 0.0.0.0 --port 8080

# Environment variables
export SCREENSHOT_PORT=8080
export SCREENSHOT_HOST=localhost
./server.exe
```

## Examples & Use Cases

### [Basic Examples](examples/basics/)
- [Single Window Capture](examples/basics/single-window.md) - Simple window screenshots with REST API, CLI, and programming examples

### [Streaming & Real-time](examples/streaming/)
- [WebSocket Live Streaming](examples/streaming/websocket-streaming.md) - Real-time window feeds with JavaScript, Python, and Node.js clients

### [Hidden & System Integration](examples/hidden-and-tray/)
- [Hidden Window Capture](examples/hidden-and-tray/hidden-window.md) - Advanced techniques for minimized and system tray applications

### [Browser Integration](examples/chrome/)
- [Chrome Tab Capture](examples/chrome/chrome-tabs.md) - Direct browser tab screenshots with Chrome DevTools integration

### [Testing & Quality Assurance](examples/testing/)
- [Visual Regression Testing](examples/testing/visual-regression.md) - Automated UI change detection with Python framework

## Advanced Configuration

### Server Configuration

The server uses a default configuration that can be customized:

```go
// Default settings
type Config struct {
    Port              int    // Default: 8080
    Host              string // Default: "localhost"
    DefaultFormat     string // Default: "png"
    Quality           int    // Default: 95
    IncludeCursor     bool   // Default: false
    LogLevel          string // Default: "info"
    ChromeTimeout     string // Default: "30s"
    StreamMaxSessions int    // Default: 10
    StreamDefaultFPS  int    // Default: 10
}
```

### Chrome DevTools Setup

For Chrome tab capture, launch Chrome with debugging enabled:

```bash
# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222

# Launch with temporary profile
chrome.exe --remote-debugging-port=9222 --user-data-dir=temp-profile
```

## Building from Source

### Prerequisites
- Go 1.21 or later
- Windows OS (for Windows API support)
- Git

### Build Instructions

```bash
# Clone the repository
git clone https://github.com/your-org/screenshot-mcp-server.git
cd screenshot-mcp-server

# Install dependencies
go mod download

# Build the server
go build -o server.exe ./cmd/server

# Run tests
go test ./...

# Start the server
./server.exe
```

### Project Structure

```
├── cmd/
│   ├── server/          # Main server application
│   └── mcpctl/          # MCP control utility
├── internal/
│   ├── screenshot/      # Screenshot capture engines
│   ├── chrome/          # Chrome DevTools integration
│   ├── window/          # Window management
│   └── ws/              # WebSocket streaming
├── pkg/
│   └── types/           # Shared data structures
└── examples/            # Usage examples and documentation
```

## Architecture

The server follows a modular architecture:

- **HTTP Server** (Gin framework) - REST API endpoints
- **WebSocket Manager** - Real-time streaming support
- **Screenshot Engine** - Core capture functionality with multiple methods
- **Chrome Manager** - Browser integration via DevTools protocol
- **MCP Handler** - JSON-RPC 2.0 support for AI agents

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/screenshot-mcp-server/issues)
- **Documentation**: See `/examples` directory for usage examples

---

**A powerful Windows screenshot server built for modern automation and AI integration.**
