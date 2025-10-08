# Single Window Capture

This example demonstrates basic window screenshot capture using different methods.

## REST API Examples

### Capture by Window Title
```bash
# Capture Notepad window
curl "http://localhost:8080/api/screenshot?method=title&target=Notepad" -o notepad.png

# Capture Calculator with JPEG format
curl "http://localhost:8080/api/screenshot?method=title&target=Calculator&format=jpeg&quality=85" -o calculator.jpg

# Capture Visual Studio Code
curl "http://localhost:8080/api/screenshot?method=title&target=Visual Studio Code&format=png" -o vscode.png
```

### Capture by Process Name
```bash
# Capture by executable name
curl "http://localhost:8080/api/screenshot?method=process&target=notepad.exe&format=png" -o notepad_process.png

# Capture browser process
curl "http://localhost:8080/api/screenshot?method=process&target=chrome.exe&format=jpeg&quality=90" -o chrome_process.jpg
```

### Capture by Window Handle
```bash
# First, get window list to find handle
curl "http://localhost:8080/api/windows" | jq '.windows[] | {title, handle}'

# Then capture by handle
curl "http://localhost:8080/api/screenshot?method=handle&target=123456&format=png" -o window_handle.png
```

## CLI Examples

```bash
# Basic CLI capture
screenshot-cli capture --method title --target "Notepad" --output notepad.png

# With quality settings
screenshot-cli capture --method title --target "Calculator" --format jpeg --quality 75 --output calc.jpg

# Verbose output
screenshot-cli capture --method title --target "Visual Studio Code" --output vscode.png --verbose
```

## PowerShell Script Example

```powershell
# capture-window.ps1
param(
    [string]$WindowTitle = "Notepad",
    [string]$OutputPath = "screenshot.png",
    [string]$Format = "png",
    [int]$Quality = 90
)

$url = "http://localhost:8080/api/screenshot?method=title&target=$WindowTitle&format=$Format&quality=$Quality"

Write-Host "Capturing window: $WindowTitle"
Invoke-WebRequest -Uri $url -OutFile $OutputPath

if (Test-Path $OutputPath) {
    Write-Host "Screenshot saved to: $OutputPath"
} else {
    Write-Host "Failed to capture screenshot"
}
```

## Python Example

```python
#!/usr/bin/env python3
import requests
import sys
from datetime import datetime

def capture_window(title, output_path=None, format='png', quality=90):
    """Capture a window screenshot"""
    
    if not output_path:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        output_path = f"screenshot_{timestamp}.{format}"
    
    params = {
        'method': 'title',
        'target': title,
        'format': format,
        'quality': quality
    }
    
    try:
        response = requests.get('http://localhost:8080/api/screenshot', 
                              params=params, stream=True)
        response.raise_for_status()
        
        with open(output_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        print(f"Screenshot saved to: {output_path}")
        return True
        
    except requests.exceptions.RequestException as e:
        print(f"Error capturing screenshot: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python capture_window.py <window_title> [output_path]")
        sys.exit(1)
    
    window_title = sys.argv[1]
    output_path = sys.argv[2] if len(sys.argv) > 2 else None
    
    capture_window(window_title, output_path)
```

## Error Handling

### Common Issues and Solutions

**Window not found:**
```bash
# Check available windows first
curl "http://localhost:8080/api/windows" | jq '.windows[] | .title'

# Use partial matching if needed
curl "http://localhost:8080/api/screenshot?method=title&target=Notepad*" -o notepad.png
```

**Permission denied:**
```bash
# Some windows require elevated permissions
# Run the server as administrator for protected windows
```

**Invalid format/quality:**
```bash
# Valid formats: png, jpeg, webp
# Quality range: 1-100 (for lossy formats only)
curl "http://localhost:8080/api/screenshot?method=title&target=Calculator&format=jpeg&quality=85" -o calc.jpg
```

## Response Formats

### Success Response
The API returns the binary image data directly. HTTP status 200 indicates success.

### Error Response
```json
{
  "error": {
    "code": "WINDOW_NOT_FOUND",
    "message": "Window with title 'NonexistentApp' not found",
    "details": {
      "method": "title",
      "target": "NonexistentApp"
    }
  }
}
```

## Best Practices

1. **Window Identification**: Use specific window titles to avoid ambiguity
2. **Format Selection**: Use PNG for quality, JPEG for smaller file sizes
3. **Error Handling**: Always check HTTP status codes and handle errors
4. **Resource Management**: Don't capture too frequently to avoid performance issues
5. **Security**: Validate window targets if accepting user input

## Next Steps

- Try [Batch Capture](batch-capture.md) for multiple windows
- Explore [Desktop Capture](desktop-monitor.md) for full screen shots
- Learn about [Hidden Window Capture](../hidden-and-tray/hidden-window.md) for minimized apps