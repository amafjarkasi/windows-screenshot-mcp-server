# PowerShell script to test WebSocket streaming functionality
# This script demonstrates how to use the WebSocket streaming feature

param(
    [Parameter()]
    [string]$ServerUrl = "localhost:8080",
    
    [Parameter(Mandatory = $true)]
    [int]$WindowId,
    
    [Parameter()]
    [int]$TestDuration = 30,
    
    [Parameter()]
    [switch]$OpenBrowser,
    
    [Parameter()]
    [switch]$RunGoClient,
    
    [Parameter()]
    [switch]$ShowLogs
)

$ErrorActionPreference = "Stop"

function Write-StatusMessage {
    param([string]$Message, [string]$Color = "White")
    Write-Host "$(Get-Date -Format 'HH:mm:ss') - $Message" -ForegroundColor $Color
}

function Test-ServerConnection {
    param([string]$Url)
    
    try {
        $response = Invoke-RestMethod -Uri "http://$Url/api/health" -Method GET -TimeoutSec 5
        return $response.status -eq "healthy"
    }
    catch {
        return $false
    }
}

function Get-AvailableWindows {
    param([string]$Url)
    
    try {
        $response = Invoke-RestMethod -Uri "http://$Url/api/windows" -Method GET -TimeoutSec 10
        return $response.windows
    }
    catch {
        Write-StatusMessage "Failed to get window list: $($_.Exception.Message)" "Red"
        return @()
    }
}

function Test-WebSocketConnection {
    param([string]$Url, [int]$WindowId)
    
    Write-StatusMessage "Testing WebSocket connection..." "Yellow"
    
    # Use Node.js or similar to test WebSocket if available
    $testScript = @"
const WebSocket = require('ws');

const ws = new WebSocket('ws://$Url/stream/$WindowId');
let frameCount = 0;

ws.on('open', function open() {
    console.log('Connected to WebSocket');
});

ws.on('message', function message(data) {
    try {
        const msg = JSON.parse(data);
        if (msg.type === 'frame') {
            frameCount++;
            console.log(`Frame ${frameCount}: ${msg.data.width}x${msg.data.height}`);
        } else {
            console.log(`Message: ${msg.type}`);
        }
    } catch (error) {
        console.log('Parse error:', error.message);
    }
});

ws.on('error', function error(err) {
    console.log('WebSocket error:', err.message);
});

ws.on('close', function close() {
    console.log(`Connection closed. Received ${frameCount} frames.`);
});

setTimeout(() => {
    ws.close();
}, 10000);
"@

    $tempScript = [System.IO.Path]::GetTempFileName() + ".js"
    Set-Content -Path $tempScript -Value $testScript
    
    try {
        if (Get-Command "node" -ErrorAction SilentlyContinue) {
            node $tempScript
        } else {
            Write-StatusMessage "Node.js not found. Skipping WebSocket test." "Yellow"
        }
    }
    finally {
        Remove-Item $tempScript -ErrorAction SilentlyContinue
    }
}

# Main execution
Write-StatusMessage "🚀 Starting WebSocket Streaming Test" "Green"
Write-StatusMessage "Server: $ServerUrl" "Cyan"
Write-StatusMessage "Window ID: $WindowId" "Cyan"
Write-StatusMessage "Test Duration: $TestDuration seconds" "Cyan"

# Test server connection
Write-StatusMessage "Checking server connection..." "Yellow"
if (-not (Test-ServerConnection -Url $ServerUrl)) {
    Write-StatusMessage "❌ Server is not running or not responding at $ServerUrl" "Red"
    Write-StatusMessage "Please start the server first:" "Yellow"
    Write-StatusMessage "  go run cmd/server/main.go" "White"
    exit 1
}

Write-StatusMessage "✅ Server is running" "Green"

# Get available windows
Write-StatusMessage "Retrieving available windows..." "Yellow"
$windows = Get-AvailableWindows -Url $ServerUrl

if ($windows.Count -eq 0) {
    Write-StatusMessage "❌ No windows found" "Red"
    exit 1
}

Write-StatusMessage "📊 Available windows:" "Green"
foreach ($window in $windows | Select-Object -First 10) {
    $status = if ($window.Visible) { "✅" } else { "❌" }
    Write-StatusMessage "  $status ID: $($window.ID) - $($window.Title)" "White"
}

# Validate window ID
$targetWindow = $windows | Where-Object { $_.ID -eq $WindowId }
if (-not $targetWindow) {
    Write-StatusMessage "❌ Window ID $WindowId not found in available windows" "Red"
    exit 1
}

Write-StatusMessage "🎯 Target window: $($targetWindow.Title)" "Green"

# Test WebSocket connection
Test-WebSocketConnection -Url $ServerUrl -WindowId $WindowId

# Open browser if requested
if ($OpenBrowser) {
    Write-StatusMessage "🌐 Opening browser viewer..." "Yellow"
    $htmlPath = Join-Path $PSScriptRoot "websocket-viewer.html"
    if (Test-Path $htmlPath) {
        $url = "file:///$($htmlPath.Replace('\', '/'))?server=$ServerUrl&windowId=$WindowId"
        Start-Process $url
        Write-StatusMessage "✅ Browser opened with streaming viewer" "Green"
    } else {
        Write-StatusMessage "❌ HTML viewer not found at $htmlPath" "Red"
    }
}

# Run Go client if requested
if ($RunGoClient) {
    Write-StatusMessage "🔧 Running Go streaming client..." "Yellow"
    $clientPath = Join-Path $PSScriptRoot "streaming-client"
    
    if (Test-Path $clientPath) {
        Push-Location $clientPath
        try {
            go mod tidy 2>$null
            $clientArgs = @(
                "run", "main.go",
                "-server", $ServerUrl,
                "-windows", $WindowId.ToString(),
                "-timeout", "$($TestDuration)s"
            )
            
            if ($ShowLogs) {
                $clientArgs += "-verbose"
            }
            
            Write-StatusMessage "Running: go $($clientArgs -join ' ')" "Cyan"
            & go @clientArgs
        }
        catch {
            Write-StatusMessage "❌ Failed to run Go client: $($_.Exception.Message)" "Red"
        }
        finally {
            Pop-Location
        }
    } else {
        Write-StatusMessage "❌ Go client not found at $clientPath" "Red"
    }
}

# Show example curl commands
Write-StatusMessage "📋 Example API calls:" "Green"
Write-StatusMessage "  Health check:" "Yellow"
Write-StatusMessage "    curl http://$ServerUrl/api/health" "White"
Write-StatusMessage "  Get windows:" "Yellow"
Write-StatusMessage "    curl http://$ServerUrl/api/windows" "White"
Write-StatusMessage "  Take screenshot:" "Yellow"
Write-StatusMessage "    curl 'http://$ServerUrl/api/screenshot?window=$WindowId&format=png' -o screenshot.png" "White"
Write-StatusMessage "  WebSocket URL:" "Yellow"
Write-StatusMessage "    ws://$ServerUrl/stream/$WindowId" "White"

Write-StatusMessage "✅ Test completed" "Green"
Write-StatusMessage "💡 To manually test streaming, open the HTML viewer or use the Go client" "Cyan"