# WebSocket Live Streaming

Real-time window streaming using WebSocket connections for live screenshot feeds.

## Basic WebSocket Connection

### JavaScript Client Example

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Screenshot Stream Viewer</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 0; 
            padding: 20px; 
            background: #f0f0f0; 
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            padding: 20px; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .controls { 
            margin-bottom: 20px; 
            padding: 15px; 
            background: #f8f9fa; 
            border-radius: 5px;
        }
        .control-group { 
            margin-bottom: 10px; 
        }
        label { 
            display: inline-block; 
            width: 120px; 
            font-weight: bold; 
        }
        input, select, button { 
            padding: 8px 12px; 
            margin: 5px; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
        }
        button { 
            background: #007bff; 
            color: white; 
            border: none; 
            cursor: pointer; 
        }
        button:hover { background: #0056b3; }
        button:disabled { background: #6c757d; cursor: not-allowed; }
        .status { 
            margin: 10px 0; 
            padding: 10px; 
            border-radius: 4px; 
            font-weight: bold;
        }
        .status.connected { background: #d4edda; color: #155724; }
        .status.disconnected { background: #f8d7da; color: #721c24; }
        .stream-container { 
            text-align: center; 
            border: 2px solid #ddd; 
            border-radius: 8px; 
            padding: 20px; 
            min-height: 400px;
        }
        .stream-image { 
            max-width: 100%; 
            max-height: 80vh; 
            border-radius: 4px;
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .stats { 
            margin-top: 15px; 
            font-family: monospace; 
            background: #f8f9fa; 
            padding: 10px; 
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Screenshot Stream Viewer</h1>
        
        <div class="controls">
            <div class="control-group">
                <label>Window ID:</label>
                <input type="text" id="windowId" value="0" placeholder="0 for desktop">
                <button onclick="listWindows()">List Windows</button>
            </div>
            
            <div class="control-group">
                <label>FPS:</label>
                <input type="range" id="fpsSlider" min="1" max="30" value="10" oninput="updateFPS()">
                <span id="fpsValue">10</span>
            </div>
            
            <div class="control-group">
                <label>Quality:</label>
                <input type="range" id="qualitySlider" min="10" max="100" value="80" oninput="updateQuality()">
                <span id="qualityValue">80</span>
            </div>
            
            <div class="control-group">
                <label>Format:</label>
                <select id="formatSelect" onchange="updateFormat()">
                    <option value="jpeg">JPEG</option>
                    <option value="png">PNG</option>
                    <option value="webp">WebP</option>
                </select>
            </div>
            
            <div class="control-group">
                <button id="connectBtn" onclick="toggleConnection()">Connect</button>
                <button onclick="takeScreenshot()">Take Screenshot</button>
                <button onclick="resetStats()">Reset Stats</button>
            </div>
        </div>
        
        <div id="status" class="status disconnected">Disconnected</div>
        
        <div class="stream-container">
            <img id="streamImage" class="stream-image" style="display: none;" alt="Stream">
            <div id="placeholder">Click Connect to start streaming</div>
        </div>
        
        <div class="stats">
            <div><strong>Statistics:</strong></div>
            <div>Frames Received: <span id="frameCount">0</span></div>
            <div>Actual FPS: <span id="actualFPS">0.0</span></div>
            <div>Data Received: <span id="dataReceived">0 KB</span></div>
            <div>Connection Time: <span id="connectionTime">0s</span></div>
            <div>Last Frame: <span id="lastFrameTime">Never</span></div>
        </div>
    </div>

    <script>
        let ws = null;
        let frameCount = 0;
        let totalDataReceived = 0;
        let connectionStartTime = null;
        let lastFrameTime = null;
        let isConnected = false;

        function toggleConnection() {
            if (isConnected) {
                disconnect();
            } else {
                connect();
            }
        }

        function connect() {
            const windowId = document.getElementById('windowId').value;
            const fps = document.getElementById('fpsSlider').value;
            const quality = document.getElementById('qualitySlider').value;
            const format = document.getElementById('formatSelect').value;
            
            const wsUrl = `ws://localhost:8080/stream/${windowId}?fps=${fps}&quality=${quality}&format=${format}`;
            
            try {
                ws = new WebSocket(wsUrl);
                
                ws.onopen = function() {
                    isConnected = true;
                    connectionStartTime = Date.now();
                    document.getElementById('status').className = 'status connected';
                    document.getElementById('status').textContent = 'Connected';
                    document.getElementById('connectBtn').textContent = 'Disconnect';
                    document.getElementById('placeholder').style.display = 'none';
                    document.getElementById('streamImage').style.display = 'block';
                    console.log('Connected to stream');
                };
                
                ws.onmessage = function(event) {
                    try {
                        const message = JSON.parse(event.data);
                        
                        if (message.type === 'frame') {
                            frameCount++;
                            lastFrameTime = Date.now();
                            totalDataReceived += event.data.length;
                            
                            document.getElementById('streamImage').src = message.data.data_url;
                            updateStats();
                        } else if (message.type === 'error') {
                            console.error('Stream error:', message.error);
                            updateStatus('Error: ' + message.error.message, 'disconnected');
                        }
                    } catch (e) {
                        console.error('Failed to parse message:', e);
                    }
                };
                
                ws.onclose = function() {
                    handleDisconnection();
                };
                
                ws.onerror = function(error) {
                    console.error('WebSocket error:', error);
                    updateStatus('Connection error', 'disconnected');
                };
                
            } catch (error) {
                console.error('Failed to connect:', error);
                updateStatus('Failed to connect: ' + error.message, 'disconnected');
            }
        }

        function disconnect() {
            if (ws) {
                ws.close();
            }
        }

        function handleDisconnection() {
            isConnected = false;
            document.getElementById('status').className = 'status disconnected';
            document.getElementById('status').textContent = 'Disconnected';
            document.getElementById('connectBtn').textContent = 'Connect';
            document.getElementById('streamImage').style.display = 'none';
            document.getElementById('placeholder').style.display = 'block';
            document.getElementById('placeholder').textContent = 'Click Connect to start streaming';
            ws = null;
        }

        function updateFPS() {
            const fps = document.getElementById('fpsSlider').value;
            document.getElementById('fpsValue').textContent = fps;
            
            if (isConnected) {
                sendCommand('update_options', { fps: parseInt(fps) });
            }
        }

        function updateQuality() {
            const quality = document.getElementById('qualitySlider').value;
            document.getElementById('qualityValue').textContent = quality;
            
            if (isConnected) {
                sendCommand('update_options', { quality: parseInt(quality) });
            }
        }

        function updateFormat() {
            const format = document.getElementById('formatSelect').value;
            
            if (isConnected) {
                sendCommand('update_options', { format: format });
            }
        }

        function sendCommand(command, options = {}) {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    command: command,
                    options: options
                }));
            }
        }

        function takeScreenshot() {
            if (isConnected) {
                sendCommand('take_screenshot');
            } else {
                alert('Not connected to stream');
            }
        }

        function listWindows() {
            fetch('http://localhost:8080/api/windows')
                .then(response => response.json())
                .then(data => {
                    let windowList = 'Available Windows:\\n\\n';
                    data.windows.forEach(window => {
                        windowList += `${window.handle}: ${window.title}\\n`;
                    });
                    alert(windowList);
                })
                .catch(error => {
                    alert('Failed to get window list: ' + error.message);
                });
        }

        function updateStats() {
            document.getElementById('frameCount').textContent = frameCount;
            
            if (connectionStartTime) {
                const connectionTime = Math.round((Date.now() - connectionStartTime) / 1000);
                document.getElementById('connectionTime').textContent = connectionTime + 's';
                
                const actualFPS = frameCount / (connectionTime || 1);
                document.getElementById('actualFPS').textContent = actualFPS.toFixed(1);
            }
            
            const dataKB = Math.round(totalDataReceived / 1024);
            document.getElementById('dataReceived').textContent = dataKB + ' KB';
            
            if (lastFrameTime) {
                document.getElementById('lastFrameTime').textContent = new Date(lastFrameTime).toLocaleTimeString();
            }
        }

        function resetStats() {
            frameCount = 0;
            totalDataReceived = 0;
            connectionStartTime = Date.now();
            lastFrameTime = null;
            updateStats();
        }

        function updateStatus(message, className) {
            document.getElementById('status').textContent = message;
            document.getElementById('status').className = 'status ' + className;
        }

        // Update stats every second
        setInterval(updateStats, 1000);
    </script>
</body>
</html>
```

## Node.js Client Example

```javascript
// stream-client.js
const WebSocket = require('ws');
const fs = require('fs');

class ScreenshotStreamClient {
    constructor(serverUrl = 'ws://localhost:8080') {
        this.serverUrl = serverUrl;
        this.ws = null;
        this.frameCount = 0;
        this.isConnected = false;
    }

    connect(windowId = '0', options = {}) {
        const params = new URLSearchParams({
            fps: options.fps || 10,
            quality: options.quality || 80,
            format: options.format || 'jpeg'
        });

        const wsUrl = `${this.serverUrl}/stream/${windowId}?${params}`;
        
        console.log(`Connecting to: ${wsUrl}`);
        this.ws = new WebSocket(wsUrl);

        this.ws.on('open', () => {
            this.isConnected = true;
            console.log('Connected to stream');
        });

        this.ws.on('message', (data) => {
            try {
                const message = JSON.parse(data.toString());
                this.handleMessage(message);
            } catch (error) {
                console.error('Failed to parse message:', error);
            }
        });

        this.ws.on('close', () => {
            this.isConnected = false;
            console.log('Stream disconnected');
        });

        this.ws.on('error', (error) => {
            console.error('WebSocket error:', error);
        });
    }

    handleMessage(message) {
        switch (message.type) {
            case 'frame':
                this.frameCount++;
                console.log(`Frame ${this.frameCount} received`);
                
                // Save frame to file (optional)
                if (this.frameCount % 30 === 0) { // Save every 30th frame
                    this.saveFrame(message.data, this.frameCount);
                }
                break;

            case 'error':
                console.error('Stream error:', message.error);
                break;

            case 'stats':
                console.log('Stream stats:', message.data);
                break;

            default:
                console.log('Unknown message type:', message.type);
        }
    }

    saveFrame(frameData, frameNumber) {
        // Extract base64 data from data URL
        const base64Data = frameData.data_url.replace(/^data:image\/\w+;base64,/, '');
        const buffer = Buffer.from(base64Data, 'base64');
        
        const filename = `frame_${frameNumber.toString().padStart(6, '0')}.${frameData.format}`;
        fs.writeFileSync(`./frames/${filename}`, buffer);
        console.log(`Saved frame: ${filename}`);
    }

    updateOptions(options) {
        if (this.isConnected) {
            this.ws.send(JSON.stringify({
                command: 'update_options',
                options: options
            }));
        }
    }

    takeScreenshot() {
        if (this.isConnected) {
            this.ws.send(JSON.stringify({
                command: 'take_screenshot'
            }));
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Example usage
const client = new ScreenshotStreamClient();

// Create frames directory
if (!fs.existsSync('./frames')) {
    fs.mkdirSync('./frames');
}

// Connect and stream
client.connect('0', {
    fps: 15,
    quality: 85,
    format: 'jpeg'
});

// Update FPS after 10 seconds
setTimeout(() => {
    client.updateOptions({ fps: 30 });
    console.log('Updated FPS to 30');
}, 10000);

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('Shutting down...');
    client.disconnect();
    process.exit(0);
});
```

## Python Client Example

```python
#!/usr/bin/env python3
import asyncio
import websockets
import json
import base64
import time
from datetime import datetime

class ScreenshotStreamClient:
    def __init__(self, server_url="ws://localhost:8080"):
        self.server_url = server_url
        self.websocket = None
        self.frame_count = 0
        self.start_time = None
        
    async def connect(self, window_id="0", fps=10, quality=80, format="jpeg"):
        uri = f"{self.server_url}/stream/{window_id}?fps={fps}&quality={quality}&format={format}"
        
        print(f"Connecting to: {uri}")
        
        try:
            self.websocket = await websockets.connect(uri)
            self.start_time = time.time()
            print("Connected to stream")
            
            # Start listening for messages
            await self.listen()
            
        except Exception as e:
            print(f"Connection failed: {e}")

    async def listen(self):
        try:
            async for message in self.websocket:
                await self.handle_message(message)
        except websockets.exceptions.ConnectionClosed:
            print("Stream disconnected")
        except Exception as e:
            print(f"Error in message handling: {e}")

    async def handle_message(self, message):
        try:
            data = json.loads(message)
            
            if data.get('type') == 'frame':
                self.frame_count += 1
                frame_data = data['data']
                
                # Calculate FPS
                elapsed = time.time() - self.start_time
                current_fps = self.frame_count / elapsed if elapsed > 0 else 0
                
                print(f"Frame {self.frame_count} | FPS: {current_fps:.1f}")
                
                # Save every 30th frame
                if self.frame_count % 30 == 0:
                    await self.save_frame(frame_data, self.frame_count)
                    
            elif data.get('type') == 'error':
                print(f"Stream error: {data['error']}")
                
        except json.JSONDecodeError as e:
            print(f"Failed to parse message: {e}")

    async def save_frame(self, frame_data, frame_number):
        try:
            # Extract base64 data from data URL
            data_url = frame_data['data_url']
            base64_data = data_url.split(',', 1)[1]
            image_data = base64.b64decode(base64_data)
            
            # Save to file
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"frame_{frame_number:06d}_{timestamp}.{frame_data['format']}"
            
            with open(f"./frames/{filename}", 'wb') as f:
                f.write(image_data)
                
            print(f"Saved frame: {filename}")
            
        except Exception as e:
            print(f"Failed to save frame: {e}")

    async def send_command(self, command, options=None):
        if self.websocket:
            message = {
                "command": command,
                "options": options or {}
            }
            await self.websocket.send(json.dumps(message))

    async def update_fps(self, fps):
        await self.send_command("update_options", {"fps": fps})
        print(f"Updated FPS to {fps}")

    async def update_quality(self, quality):
        await self.send_command("update_options", {"quality": quality})
        print(f"Updated quality to {quality}")

    async def take_screenshot(self):
        await self.send_command("take_screenshot")
        print("Screenshot command sent")

# Example usage
async def main():
    client = ScreenshotStreamClient()
    
    # Create frames directory
    import os
    os.makedirs("./frames", exist_ok=True)
    
    # Connect and stream
    await client.connect(
        window_id="0",  # Desktop
        fps=15,
        quality=85,
        format="jpeg"
    )

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Stream stopped by user")
```

## Advanced Features

### Dynamic Quality Adjustment

```javascript
// Automatically adjust quality based on bandwidth
class AdaptiveStreaming {
    constructor(websocket) {
        this.ws = websocket;
        this.lastFrameTime = Date.now();
        this.frameInterval = 100; // Expected interval in ms
        this.currentQuality = 80;
        this.targetFPS = 15;
    }
    
    onFrame() {
        const now = Date.now();
        const actualInterval = now - this.lastFrameTime;
        this.lastFrameTime = now;
        
        // Adjust quality based on performance
        if (actualInterval > this.frameInterval * 1.5) {
            // Slow performance, reduce quality
            this.currentQuality = Math.max(30, this.currentQuality - 10);
            this.updateQuality();
        } else if (actualInterval < this.frameInterval * 0.8) {
            // Good performance, can increase quality
            this.currentQuality = Math.min(95, this.currentQuality + 5);
            this.updateQuality();
        }
    }
    
    updateQuality() {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                command: 'update_options',
                options: { quality: this.currentQuality }
            }));
        }
    }
}
```

### Multi-Window Streaming

```javascript
class MultiWindowStreamer {
    constructor() {
        this.streams = new Map();
    }
    
    addStream(windowId, containerId, options = {}) {
        const ws = new WebSocket(
            `ws://localhost:8080/stream/${windowId}?fps=${options.fps || 10}&quality=${options.quality || 80}`
        );
        
        const container = document.getElementById(containerId);
        const img = document.createElement('img');
        img.style.maxWidth = '100%';
        container.appendChild(img);
        
        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            if (message.type === 'frame') {
                img.src = message.data.data_url;
            }
        };
        
        this.streams.set(windowId, { ws, img, container });
    }
    
    removeStream(windowId) {
        const stream = this.streams.get(windowId);
        if (stream) {
            stream.ws.close();
            stream.container.removeChild(stream.img);
            this.streams.delete(windowId);
        }
    }
    
    closeAll() {
        for (const [windowId, stream] of this.streams) {
            stream.ws.close();
        }
        this.streams.clear();
    }
}
```

## Error Handling

### Connection Recovery

```javascript
class ReliableStreamClient {
    constructor(wsUrl, options = {}) {
        this.wsUrl = wsUrl;
        this.options = options;
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
        this.isIntentionalClose = false;
    }
    
    connect() {
        try {
            this.ws = new WebSocket(this.wsUrl);
            
            this.ws.onopen = () => {
                console.log('Connected to stream');
                this.reconnectDelay = 1000;
                this.reconnectAttempts = 0;
            };
            
            this.ws.onmessage = (event) => {
                this.handleMessage(event);
            };
            
            this.ws.onclose = (event) => {
                if (!this.isIntentionalClose) {
                    this.attemptReconnect();
                }
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
            
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.attemptReconnect();
        }
    }
    
    attemptReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            return;
        }
        
        this.reconnectAttempts++;
        console.log(`Reconnection attempt ${this.reconnectAttempts} in ${this.reconnectDelay}ms`);
        
        setTimeout(() => {
            this.connect();
        }, this.reconnectDelay);
        
        // Exponential backoff
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
    }
    
    disconnect() {
        this.isIntentionalClose = true;
        if (this.ws) {
            this.ws.close();
        }
    }
}
```

## Performance Optimization

### Frame Rate Control

```javascript
class FrameRateController {
    constructor(targetFPS = 15) {
        this.targetFPS = targetFPS;
        this.frameInterval = 1000 / targetFPS;
        this.lastFrameTime = 0;
        this.frameBuffer = [];
    }
    
    shouldProcessFrame() {
        const now = Date.now();
        if (now - this.lastFrameTime >= this.frameInterval) {
            this.lastFrameTime = now;
            return true;
        }
        return false;
    }
    
    onFrame(frameData) {
        if (this.shouldProcessFrame()) {
            // Process frame immediately
            this.displayFrame(frameData);
            
            // Clear buffer
            this.frameBuffer = [];
        } else {
            // Buffer frame for later
            this.frameBuffer.push(frameData);
            
            // Keep only latest frame in buffer
            if (this.frameBuffer.length > 1) {
                this.frameBuffer.shift();
            }
        }
    }
}
```

## Best Practices

1. **Connection Management**: Always handle reconnection scenarios
2. **Frame Rate**: Balance FPS with system performance and bandwidth
3. **Quality Settings**: Adjust quality based on use case and network conditions
4. **Error Handling**: Implement robust error handling and user feedback
5. **Resource Cleanup**: Properly close WebSocket connections
6. **Security**: Validate window IDs and parameters from user input

## Next Steps

- Explore [Dynamic Quality Control](dynamic-quality.md)
- Learn about [Chrome Tab Capture](../chrome/chrome-tabs.md) for browser streaming
- Try [Multi-Monitor Streaming](../basics/desktop-monitor.md) for multiple displays