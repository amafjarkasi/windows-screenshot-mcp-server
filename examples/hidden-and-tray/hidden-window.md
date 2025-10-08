# Hidden Window Capture

Advanced techniques for capturing minimized, hidden, and system tray applications.

## Overview

Hidden window capture is one of the most challenging aspects of screenshot automation. This server implements multiple fallback strategies to capture windows in various states:

- **Minimized windows**: Standard windows that are collapsed to taskbar
- **Hidden windows**: Windows with `SW_HIDE` state or zero opacity
- **System tray applications**: Apps running in the notification area
- **Cloaked windows**: UWP applications that are "cloaked" by DWM
- **Background services**: Headless applications with UI components

## Capture Methods

### 1. DWM Thumbnail (Recommended)

The fastest and most reliable method for Windows 7+ systems.

```bash
# Basic hidden window capture
curl "http://localhost:8080/api/screenshot?method=process&target=notepad.exe&allow_hidden=true" -o hidden_notepad.png

# Force DWM thumbnail method
curl "http://localhost:8080/api/screenshot?method=dwm_thumbnail&target=123456&allow_cloaked=true" -o dwm_capture.png
```

### 2. PrintWindow API

Compatible with most Windows applications, works well for minimized windows.

```bash
# Use PrintWindow method
curl "http://localhost:8080/api/screenshot?method=print_window&target=notepad.exe&restore_window=false" -o print_capture.png
```

### 3. Stealth Window Restoration

Temporarily restores windows without user disruption.

```bash
# Stealth restore for stubborn applications
curl "http://localhost:8080/api/screenshot?method=stealth_restore&target=calculator.exe&restore_timeout=1000" -o stealth_capture.png
```

## REST API Examples

### Basic Hidden Capture

```bash
# Capture minimized Notepad
curl "http://localhost:8080/api/screenshot?method=process&target=notepad.exe&allow_hidden=true" -o minimized_notepad.png

# Capture system tray app
curl "http://localhost:8080/api/screenshot?method=tray&target=explorer.exe&detect_tray=true" -o tray_explorer.png

# Capture UWP cloaked window
curl "http://localhost:8080/api/screenshot?method=process&target=Calculator.exe&allow_cloaked=true" -o uwp_calc.png
```

### Advanced Options

```bash
# Multiple fallback methods
curl "http://localhost:8080/api/screenshot?method=auto&target=notepad.exe&fallback_methods=dwm_thumbnail,print_window,wm_print&timeout=5000" -o robust_capture.png

# Process-based discovery
curl "http://localhost:8080/api/screenshot?method=process&target=chrome.exe&enumerate_all=true&select_main=true" -o main_chrome.png

# Custom region from hidden window
curl "http://localhost:8080/api/screenshot?method=process&target=notepad.exe&region=10,10,400,300&allow_hidden=true" -o region_capture.png
```

## Discovery APIs

### Find Hidden Windows

```bash
# List all hidden windows
curl "http://localhost:8080/api/windows/hidden" | jq '.windows[] | {title, process, state}'

# Find windows by process
curl "http://localhost:8080/api/windows?process=notepad.exe&include_hidden=true" | jq '.windows'

# System tray applications
curl "http://localhost:8080/api/windows/tray" | jq '.tray_apps'
```

## CLI Examples

### Basic CLI Usage

```bash
# Discover hidden applications
screenshot-cli discover --hidden --output hidden_apps.json

# Capture specific hidden window
screenshot-cli capture --method process --target "notepad.exe" --allow-hidden --output hidden.png

# Batch capture all hidden windows
screenshot-cli batch --input hidden_apps.json --method auto --allow-hidden --output-dir ./hidden_screenshots/
```

### Advanced CLI Operations

```bash
# Test fallback methods
screenshot-cli test-fallbacks --target "Calculator.exe" --methods "dwm_thumbnail,print_window,stealth_restore"

# Monitor system tray changes
screenshot-cli monitor-tray --interval 30s --output-dir ./tray_monitoring/

# Process enumeration
screenshot-cli enum-process --name "chrome.exe" --include-child --output chrome_windows.json
```

## Programming Examples

### Python Implementation

```python
#!/usr/bin/env python3
import requests
import json
import time
from typing import List, Dict, Optional

class HiddenWindowCapture:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        
    def discover_hidden_windows(self) -> List[Dict]:
        """Discover all hidden windows"""
        response = self.session.get(f"{self.base_url}/api/windows/hidden")
        response.raise_for_status()
        return response.json().get('windows', [])
    
    def find_by_process(self, process_name: str, include_hidden: bool = True) -> List[Dict]:
        """Find windows by process name"""
        params = {
            'process': process_name,
            'include_hidden': include_hidden
        }
        response = self.session.get(f"{self.base_url}/api/windows", params=params)
        response.raise_for_status()
        return response.json().get('windows', [])
    
    def capture_hidden_window(self, target: str, method: str = "auto", **kwargs) -> bytes:
        """Capture a hidden window"""
        params = {
            'method': method,
            'target': target,
            'allow_hidden': True,
            **kwargs
        }
        
        response = self.session.get(f"{self.base_url}/api/screenshot", params=params)
        response.raise_for_status()
        return response.content
    
    def test_capture_methods(self, target: str) -> Dict[str, Dict]:
        """Test different capture methods for a target"""
        methods = ['dwm_thumbnail', 'print_window', 'wm_print', 'stealth_restore']
        results = {}
        
        for method in methods:
            try:
                start_time = time.time()
                image_data = self.capture_hidden_window(target, method=method)
                duration = time.time() - start_time
                
                results[method] = {
                    'success': True,
                    'size': len(image_data),
                    'duration': duration,
                    'error': None
                }
                
                # Save test result
                with open(f"test_{method}_{target}.png", 'wb') as f:
                    f.write(image_data)
                    
            except Exception as e:
                results[method] = {
                    'success': False,
                    'size': 0,
                    'duration': 0,
                    'error': str(e)
                }
        
        return results
    
    def monitor_tray_applications(self, interval: int = 30) -> None:
        """Monitor system tray applications"""
        print("Monitoring system tray applications...")
        
        previous_apps = set()
        
        while True:
            try:
                response = self.session.get(f"{self.base_url}/api/windows/tray")
                tray_apps = response.json().get('tray_apps', [])
                
                current_apps = {app['title'] for app in tray_apps}
                
                # Detect changes
                new_apps = current_apps - previous_apps
                removed_apps = previous_apps - current_apps
                
                if new_apps:
                    print(f"New tray apps: {', '.join(new_apps)}")
                    for app_title in new_apps:
                        try:
                            self.capture_tray_app(app_title)
                        except Exception as e:
                            print(f"Failed to capture {app_title}: {e}")
                
                if removed_apps:
                    print(f"Removed tray apps: {', '.join(removed_apps)}")
                
                previous_apps = current_apps
                time.sleep(interval)
                
            except KeyboardInterrupt:
                print("Monitoring stopped by user")
                break
            except Exception as e:
                print(f"Monitoring error: {e}")
                time.sleep(interval)
    
    def capture_tray_app(self, app_title: str) -> bytes:
        """Capture a system tray application"""
        return self.capture_hidden_window(
            target=app_title,
            method='tray',
            detect_tray=True
        )

# Example usage
if __name__ == "__main__":
    capture = HiddenWindowCapture()
    
    # Discover all hidden windows
    print("Discovering hidden windows...")
    hidden_windows = capture.discover_hidden_windows()
    print(f"Found {len(hidden_windows)} hidden windows")
    
    for window in hidden_windows[:5]:  # Show first 5
        print(f"  - {window.get('title', 'Untitled')} ({window.get('process', 'Unknown')})")
    
    # Test capture methods for Notepad
    if hidden_windows:
        target_process = "notepad.exe"
        print(f"\nTesting capture methods for {target_process}...")
        results = capture.test_capture_methods(target_process)
        
        for method, result in results.items():
            status = "✅" if result['success'] else "❌"
            print(f"  {status} {method}: {result['duration']:.2f}s ({result['size']} bytes)")
            if result['error']:
                print(f"    Error: {result['error']}")
    
    # Start tray monitoring (comment out for non-interactive use)
    # capture.monitor_tray_applications(interval=10)
```

### PowerShell Implementation

```powershell
# HiddenWindowCapture.ps1

class HiddenWindowCapture {
    [string]$BaseUrl
    
    HiddenWindowCapture([string]$BaseUrl = "http://localhost:8080") {
        $this.BaseUrl = $BaseUrl
    }
    
    [PSObject[]] DiscoverHiddenWindows() {
        try {
            $response = Invoke-RestMethod -Uri "$($this.BaseUrl)/api/windows/hidden" -Method Get
            return $response.windows
        }
        catch {
            Write-Error "Failed to discover hidden windows: $_"
            return @()
        }
    }
    
    [PSObject[]] FindByProcess([string]$ProcessName) {
        try {
            $params = @{
                Uri = "$($this.BaseUrl)/api/windows"
                Method = "Get"
                Body = @{
                    process = $ProcessName
                    include_hidden = $true
                }
            }
            $response = Invoke-RestMethod @params
            return $response.windows
        }
        catch {
            Write-Error "Failed to find windows for process ${ProcessName}: $_"
            return @()
        }
    }
    
    [void] CaptureHiddenWindow([string]$Target, [string]$OutputPath, [string]$Method = "auto") {
        try {
            $params = @{
                method = $Method
                target = $Target
                allow_hidden = $true
            }
            
            $uri = "$($this.BaseUrl)/api/screenshot"
            $queryString = ($params.GetEnumerator() | ForEach-Object { "$($_.Key)=$($_.Value)" }) -join "&"
            $fullUri = "${uri}?${queryString}"
            
            Write-Host "Capturing hidden window: $Target using method: $Method"
            Invoke-WebRequest -Uri $fullUri -OutFile $OutputPath
            
            if (Test-Path $OutputPath) {
                $fileInfo = Get-Item $OutputPath
                Write-Host "✅ Screenshot saved: $OutputPath ($($fileInfo.Length) bytes)"
            }
        }
        catch {
            Write-Error "Failed to capture hidden window ${Target}: $_"
        }
    }
    
    [hashtable] TestCaptureMethods([string]$Target) {
        $methods = @("dwm_thumbnail", "print_window", "wm_print", "stealth_restore")
        $results = @{}
        
        Write-Host "Testing capture methods for: $Target"
        
        foreach ($method in $methods) {
            Write-Host "  Testing $method..." -NoNewline
            
            $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
            $outputPath = "test_${method}_${Target}.png"
            
            try {
                $this.CaptureHiddenWindow($Target, $outputPath, $method)
                $stopwatch.Stop()
                
                $fileSize = if (Test-Path $outputPath) { (Get-Item $outputPath).Length } else { 0 }
                
                $results[$method] = @{
                    Success = $true
                    Duration = $stopwatch.ElapsedMilliseconds
                    Size = $fileSize
                    Error = $null
                }
                
                Write-Host " ✅ ($($stopwatch.ElapsedMilliseconds)ms, $fileSize bytes)"
            }
            catch {
                $stopwatch.Stop()
                $results[$method] = @{
                    Success = $false
                    Duration = $stopwatch.ElapsedMilliseconds
                    Size = 0
                    Error = $_.Exception.Message
                }
                
                Write-Host " ❌ ($($_.Exception.Message))"
            }
        }
        
        return $results
    }
    
    [void] MonitorTrayApplications([int]$IntervalSeconds = 30) {
        Write-Host "Starting system tray monitoring (interval: ${IntervalSeconds}s)"
        Write-Host "Press Ctrl+C to stop"
        
        $previousApps = @()
        
        try {
            while ($true) {
                $response = Invoke-RestMethod -Uri "$($this.BaseUrl)/api/windows/tray" -Method Get
                $currentApps = $response.tray_apps | ForEach-Object { $_.title }
                
                $newApps = $currentApps | Where-Object { $_ -notin $previousApps }
                $removedApps = $previousApps | Where-Object { $_ -notin $currentApps }
                
                if ($newApps) {
                    Write-Host "$(Get-Date -Format 'HH:mm:ss') - New tray apps: $($newApps -join ', ')" -ForegroundColor Green
                    foreach ($app in $newApps) {
                        try {
                            $outputPath = "tray_${app}_$(Get-Date -Format 'yyyyMMdd_HHmmss').png"
                            $this.CaptureHiddenWindow($app, $outputPath, "tray")
                        }
                        catch {
                            Write-Host "  Failed to capture $app" -ForegroundColor Red
                        }
                    }
                }
                
                if ($removedApps) {
                    Write-Host "$(Get-Date -Format 'HH:mm:ss') - Removed tray apps: $($removedApps -join ', ')" -ForegroundColor Yellow
                }
                
                $previousApps = $currentApps
                Start-Sleep -Seconds $IntervalSeconds
            }
        }
        catch [System.Management.Automation.RuntimeException] {
            Write-Host "`nTray monitoring stopped by user" -ForegroundColor Yellow
        }
        catch {
            Write-Error "Monitoring error: $_"
        }
    }
}

# Example usage
$capture = [HiddenWindowCapture]::new()

Write-Host "🔍 Discovering hidden windows..."
$hiddenWindows = $capture.DiscoverHiddenWindows()
Write-Host "Found $($hiddenWindows.Count) hidden windows:"

$hiddenWindows | Select-Object -First 10 | ForEach-Object {
    $title = if ($_.title) { $_.title } else { "Untitled" }
    $process = if ($_.process) { $_.process } else { "Unknown" }
    Write-Host "  • $title ($process)" -ForegroundColor Cyan
}

# Test capture methods
if ($hiddenWindows.Count -gt 0) {
    Write-Host "`n🧪 Testing capture methods..."
    $testTarget = "notepad.exe"
    $results = $capture.TestCaptureMethods($testTarget)
    
    Write-Host "`n📊 Results Summary:"
    $results.GetEnumerator() | Sort-Object Key | ForEach-Object {
        $status = if ($_.Value.Success) { "✅" } else { "❌" }
        Write-Host "  $status $($_.Key): $($_.Value.Duration)ms ($($_.Value.Size) bytes)"
    }
}

# Uncomment to start tray monitoring
# Write-Host "`n👀 Starting tray monitoring..."
# $capture.MonitorTrayApplications(10)
```

### Node.js Implementation

```javascript
// hidden-capture.js
const axios = require('axios');
const fs = require('fs');
const path = require('path');

class HiddenWindowCapture {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.axios = axios.create({ baseURL: baseUrl });
    }

    async discoverHiddenWindows() {
        try {
            const response = await this.axios.get('/api/windows/hidden');
            return response.data.windows || [];
        } catch (error) {
            console.error('Failed to discover hidden windows:', error.message);
            return [];
        }
    }

    async findByProcess(processName, includeHidden = true) {
        try {
            const response = await this.axios.get('/api/windows', {
                params: { process: processName, include_hidden: includeHidden }
            });
            return response.data.windows || [];
        } catch (error) {
            console.error(`Failed to find windows for ${processName}:`, error.message);
            return [];
        }
    }

    async captureHiddenWindow(target, method = 'auto', options = {}) {
        try {
            const params = {
                method,
                target,
                allow_hidden: true,
                ...options
            };

            const response = await this.axios.get('/api/screenshot', {
                params,
                responseType: 'arraybuffer'
            });

            return Buffer.from(response.data);
        } catch (error) {
            throw new Error(`Failed to capture ${target}: ${error.message}`);
        }
    }

    async testCaptureMethods(target) {
        const methods = ['dwm_thumbnail', 'print_window', 'wm_print', 'stealth_restore'];
        const results = {};

        console.log(`🧪 Testing capture methods for: ${target}`);

        for (const method of methods) {
            process.stdout.write(`  Testing ${method}... `);
            
            const startTime = Date.now();
            
            try {
                const imageData = await this.captureHiddenWindow(target, method);
                const duration = Date.now() - startTime;

                // Save test result
                const filename = `test_${method}_${target.replace(/[^a-z0-9]/gi, '_')}.png`;
                fs.writeFileSync(filename, imageData);

                results[method] = {
                    success: true,
                    duration,
                    size: imageData.length,
                    error: null,
                    filename
                };

                console.log(`✅ (${duration}ms, ${imageData.length} bytes)`);

            } catch (error) {
                const duration = Date.now() - startTime;
                
                results[method] = {
                    success: false,
                    duration,
                    size: 0,
                    error: error.message,
                    filename: null
                };

                console.log(`❌ (${error.message})`);
            }
        }

        return results;
    }

    async monitorTrayApplications(intervalSeconds = 30) {
        console.log(`👀 Starting tray monitoring (interval: ${intervalSeconds}s)`);
        console.log('Press Ctrl+C to stop');

        let previousApps = new Set();

        const monitor = async () => {
            try {
                const response = await this.axios.get('/api/windows/tray');
                const currentApps = new Set(
                    response.data.tray_apps.map(app => app.title)
                );

                // Detect changes
                const newApps = [...currentApps].filter(app => !previousApps.has(app));
                const removedApps = [...previousApps].filter(app => !currentApps.has(app));

                if (newApps.length > 0) {
                    console.log(`${new Date().toLocaleTimeString()} - New tray apps: ${newApps.join(', ')}`);
                    
                    for (const app of newApps) {
                        try {
                            const imageData = await this.captureHiddenWindow(app, 'tray', { detect_tray: true });
                            const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
                            const filename = `tray_${app.replace(/[^a-z0-9]/gi, '_')}_${timestamp}.png`;
                            fs.writeFileSync(filename, imageData);
                            console.log(`  📸 Captured: ${filename}`);
                        } catch (error) {
                            console.log(`  ❌ Failed to capture ${app}`);
                        }
                    }
                }

                if (removedApps.length > 0) {
                    console.log(`${new Date().toLocaleTimeString()} - Removed tray apps: ${removedApps.join(', ')}`);
                }

                previousApps = currentApps;

            } catch (error) {
                console.error('Monitoring error:', error.message);
            }
        };

        // Initial check
        await monitor();

        // Set up interval
        const intervalId = setInterval(monitor, intervalSeconds * 1000);

        // Handle graceful shutdown
        process.on('SIGINT', () => {
            console.log('\n🛑 Tray monitoring stopped by user');
            clearInterval(intervalId);
            process.exit(0);
        });
    }
}

// Example usage
async function main() {
    const capture = new HiddenWindowCapture();

    try {
        // Discover hidden windows
        console.log('🔍 Discovering hidden windows...');
        const hiddenWindows = await capture.discoverHiddenWindows();
        console.log(`Found ${hiddenWindows.length} hidden windows:`);

        hiddenWindows.slice(0, 10).forEach(window => {
            const title = window.title || 'Untitled';
            const process = window.process || 'Unknown';
            console.log(`  • ${title} (${process})`);
        });

        // Test capture methods
        if (hiddenWindows.length > 0) {
            console.log('\n🧪 Testing capture methods...');
            const results = await capture.testCaptureMethods('notepad.exe');

            console.log('\n📊 Results Summary:');
            Object.entries(results)
                .sort(([a], [b]) => a.localeCompare(b))
                .forEach(([method, result]) => {
                    const status = result.success ? '✅' : '❌';
                    console.log(`  ${status} ${method}: ${result.duration}ms (${result.size} bytes)`);
                });
        }

        // Start tray monitoring (uncomment to enable)
        // await capture.monitorTrayApplications(10);

    } catch (error) {
        console.error('Error in main:', error.message);
    }
}

// Run if called directly
if (require.main === module) {
    main().catch(console.error);
}

module.exports = HiddenWindowCapture;
```

## Troubleshooting

### Common Issues

1. **Access Denied Errors**
   - Run the server with administrator privileges
   - Some system processes require elevated access

2. **Empty Screenshots**
   - Try different capture methods
   - Check if the window actually has content
   - Verify the window isn't completely transparent

3. **DWM Thumbnail Failures**
   - Ensure DWM is enabled (Windows Aero)
   - Some applications don't support DWM thumbnails
   - Try PrintWindow as fallback

4. **Process Not Found**
   - Check exact process name with Task Manager
   - Include file extension (.exe)
   - Some processes have multiple instances

### Advanced Diagnostics

```bash
# Test server capabilities
curl "http://localhost:8080/api/capabilities" | jq '.hidden_capture'

# Get detailed window information
curl "http://localhost:8080/api/windows/123456/details" | jq '.capture_methods'

# Server-side logging
curl "http://localhost:8080/api/debug/capture-log?target=notepad.exe&method=auto"
```

## Best Practices

1. **Method Selection**: Start with "auto" method, specify others only when needed
2. **Error Handling**: Always implement fallback strategies
3. **Performance**: DWM thumbnails are fastest, stealth restore is slowest
4. **Permissions**: Run with appropriate privileges for target applications
5. **Resource Management**: Don't keep unnecessary windows restored
6. **Testing**: Test capture methods during development, not production

## Next Steps

- Learn about [System Tray Applications](tray-app.md) for notification area capture
- Explore [Chrome Tab Capture](../chrome/chrome-tabs.md) for browser automation
- Try [Visual Regression Testing](../testing/visual-regression.md) with hidden windows