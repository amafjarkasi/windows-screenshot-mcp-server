# 📸 Screenshot MCP Server

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Windows Support](https://img.shields.io/badge/Windows-10/11-0078D4?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)
[![WebSocket](https://img.shields.io/badge/WebSocket-Real_Time-FF6B6B?style=for-the-badge)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
[![MCP Protocol](https://img.shields.io/badge/MCP-JSON--RPC_2.0-4ECDC4?style=for-the-badge)](https://spec.modelcontextprotocol.io/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**A powerful, production-ready MCP (Model Context Protocol) server for Windows screenshot automation with real-time WebSocket streaming capabilities.**

## ✨ Features

### 🎯 **Core Capabilities**
- **Native Windows Integration** - Direct GDI32/User32 API calls for maximum performance
- **Real-time WebSocket Streaming** - Live screenshot streaming at up to 60 FPS
- **Chrome DevTools Protocol** - Seamless browser tab capture and management
- **Multi-format Support** - PNG, JPEG, BMP, WebP with quality controls
- **Advanced Window Management** - Multi-monitor, DPI-aware, window enumeration

### 🔌 **API Interfaces**
- **HTTP REST API** - Standard HTTP endpoints for integration
- **WebSocket Streaming API** - Real-time bidirectional communication
- **MCP JSON-RPC 2.0** - Full Model Context Protocol compliance
- **Command Line Interface** - Comprehensive CLI tools

### 🚀 **Production Features**
- **High Performance** - Native Windows APIs with minimal overhead
- **Error Handling** - Robust error recovery and graceful degradation
- **Session Management** - Concurrent streaming sessions with statistics
- **Security** - CORS support, input validation, rate limiting ready
- **Observability** - Structured logging, metrics, health checks

## 🚀 Quick Start

### Prerequisites
- **Windows 10/11** (required for native Windows API support)
- **Go 1.22+** ([Download](https://golang.org/dl/))
- **Git** ([Download](https://git-scm.com/downloads))

### Installation

```bash
# Clone the repository
git clone https://github.com/amafjarkasi/windows-screenshot-mcp-server.git
cd windows-screenshot-mcp-server

# Download dependencies
make deps

# Build all components
make build

# Run the server
make run-server
```

**Server will be available at:** `http://localhost:8080`

### 🌐 WebSocket Streaming Demo

1. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

2. **Open the HTML viewer:**
   ```bash
   # Open examples/websocket-viewer.html in your browser
   start examples/websocket-viewer.html
   ```

3. **Or use the Go client:**
   ```bash
   cd examples/streaming-client
   go run main.go -windows 0 -server localhost:8080
   ```

4. **Or test with PowerShell:**
   ```powershell
   .\examples\test-streaming.ps1 -WindowId 0 -OpenBrowser
   ```

## 🏗️ Architecture

```
screenshot-mcp-server/
├── cmd/                    # Application entry points
│   ├── server/            # HTTP/WebSocket server
│   └── client/            # CLI client tools
├── internal/               # Private application code
│   ├── screenshot/        # Core screenshot engine
│   ├── chrome/            # Chrome DevTools integration
│   └── ws/                # WebSocket streaming
├── pkg/                   # Public library code
│   └── types/             # Shared types and interfaces
├── examples/              # Usage examples
│   ├── websocket-viewer.html    # Interactive web viewer
│   ├── streaming-client/        # Go client example
│   └── test-streaming.ps1       # PowerShell test script
└── test/                  # Integration tests
```

## 📖 API Documentation

### HTTP REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Server health check |
| `/api/windows` | GET | List available windows |
| `/api/screenshot` | GET | Take single screenshot |
| `/v1/stream/status` | GET | Streaming statistics |

### WebSocket Streaming API

```javascript
// Connect to streaming endpoint
const ws = new WebSocket('ws://localhost:8080/stream/0?fps=10&quality=80&format=jpeg');

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    if (message.type === 'frame') {
        // Display the frame
        img.src = message.data.data_url;
    }
};

// Control streaming
ws.send(JSON.stringify({
    command: 'update_options',
    options: { fps: 15, quality: 90, format: 'png' }
}));
```

### MCP JSON-RPC 2.0 API

```bash
# Get streaming status
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "stream.status", "id": 1}'

# Take screenshot
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "screenshot.capture", "params": {"target": "0", "method": "handle"}, "id": 1}'
```

## 🔧 Configuration

### Server Configuration
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

### Environment Variables
```bash
export SCREENSHOT_PORT=8080
export SCREENSHOT_LOG_LEVEL=info
export SCREENSHOT_MAX_SESSIONS=10
```

## 💡 Usage Examples

### Basic Screenshot
```bash
# Take screenshot of window by handle/ID
curl "http://localhost:8080/api/screenshot?window=123456&format=png" -o screenshot.png

# Using CLI
./bin/screenshot-server.exe screenshot --window 123456 --format png --output screenshot.png
```

### WebSocket Streaming
```bash
# Desktop streaming (window ID 0)
ws://localhost:8080/stream/0

# Specific window with options
ws://localhost:8080/stream/123456?fps=15&quality=90&format=jpeg
```

### Chrome Tab Capture
```bash
# List Chrome instances
curl http://localhost:8080/v1/chrome/instances

# Capture specific tab
curl -X POST http://localhost:8080/v1/chrome/tabs/TAB_ID/screenshot
```

## 🧪 Testing

### Run All Tests
```bash
# Unit tests only
make test-unit

# Integration tests (starts server automatically)
make test-integration

# All tests
make test-all-tests

# With coverage
make test-coverage

# Benchmarks
make bench
```

### Manual Testing
```bash
# PowerShell comprehensive test
.\examples\test-streaming.ps1 -WindowId 0 -OpenBrowser -RunGoClient

# Quick functionality test
make quick-test
```

## 📊 Performance

### Benchmarks
- **Screenshot Capture:** ~10-50ms (depends on window size)
- **WebSocket Streaming:** Up to 60 FPS
- **Memory Usage:** ~20-50MB baseline
- **Concurrent Sessions:** 10+ (configurable)

### Optimization Tips
- **High FPS:** Use JPEG format with 70-80% quality
- **High Quality:** Use PNG format at 5-10 FPS
- **Bandwidth:** ~500KB-2MB/sec per stream (depends on resolution)

## 🔧 Development

### Setup Development Environment
```bash
# Install development tools
make dev-setup

# Format code
make fmt

# Run linters
make lint

# Generate documentation
make docs
```

### Build for All Platforms
```bash
# Build for Windows, Linux, macOS
make build-all

# Release build with optimizations
make release
```

## 📦 Examples

### HTML WebSocket Viewer
Interactive web interface for real-time streaming with controls for:
- FPS adjustment (1-60)
- Quality control (1-100%)
- Format selection (PNG/JPEG/WebP)
- Connection statistics

### Go Streaming Client
Programmatic client example with features:
- Command-line interface
- Real-time option updates
- Graceful shutdown handling
- Structured logging

### PowerShell Test Suite
Comprehensive testing script:
- Server connectivity validation
- Window enumeration
- WebSocket connection testing
- Integration with examples

## 🛠️ Advanced Features

### Multi-Monitor Support
```go
// Capture from specific monitor
options := &types.CaptureOptions{
    MonitorIndex: 1,
    Region: &types.Rectangle{X: 0, Y: 0, Width: 1920, Height: 1080},
}
```

### Custom Screenshot Processing
```go
// Resize and compress
processor := screenshot.NewImageProcessor()
resized, _ := processor.Resize(buffer, 800, 600)
encoded, _ := processor.Encode(resized, types.FormatJPEG, 80)
```

### WebSocket Session Management
```javascript
// Monitor streaming statistics
fetch('/v1/stream/status')
  .then(r => r.json())
  .then(stats => {
    console.log(`Active sessions: ${stats.active_sessions}/${stats.max_sessions}`);
    console.log(`Total frames: ${stats.total_frames}`);
  });
```

## 🚨 Troubleshooting

### Common Issues

**Window Capture Fails:**
```bash
# Some windows require admin privileges
# Run server as administrator for protected windows
```

**WebSocket Connection Drops:**
```bash
# Check firewall settings
# Verify window ID exists: curl http://localhost:8080/api/windows
```

**High CPU Usage:**
```bash
# Reduce FPS: ws://localhost:8080/stream/ID?fps=5
# Lower quality: &quality=50
# Use JPEG format: &format=jpeg
```

### Debug Mode
```bash
# Run with verbose logging
go run cmd/server/main.go --log-level debug

# Enable race detection
go run -race cmd/server/main.go
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch:** `git checkout -b feature/amazing-feature`
3. **Commit changes:** `git commit -m 'Add amazing feature'`
4. **Push to branch:** `git push origin feature/amazing-feature`
5. **Create Pull Request**

### Development Guidelines
- Follow Go best practices
- Add tests for new features
- Update documentation
- Use conventional commits

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Windows API Documentation** - Microsoft Developer Network
- **Chrome DevTools Protocol** - Google Chrome Team
- **WebSocket Protocol** - RFC 6455 Specification
- **MCP Specification** - Model Context Protocol Working Group

## 📞 Support

- **Issues:** [GitHub Issues](https://github.com/amafjarkasi/windows-screenshot-mcp-server/issues)
- **Documentation:** [Wiki](https://github.com/amafjarkasi/windows-screenshot-mcp-server/wiki)
- **Discussions:** [GitHub Discussions](https://github.com/amafjarkasi/windows-screenshot-mcp-server/discussions)

---

**⭐ Star this repository if you find it useful!**

Built with ❤️ for the developer community.