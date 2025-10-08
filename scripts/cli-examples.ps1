# PowerShell CLI Examples for Screenshot MCP Server
# Usage examples for various screenshot scenarios

param(
    [string]$ServerUrl = "http://localhost:8080",
    [string]$OutputDir = "./screenshots"
)

# Ensure output directory exists
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
    Write-Host "Created output directory: $OutputDir" -ForegroundColor Green
}

Write-Host "Screenshot MCP Server CLI Examples" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan

# Test server connectivity
function Test-ServerConnection {
    try {
        $response = Invoke-RestMethod -Uri "$ServerUrl/health" -Method Get -TimeoutSec 5
        Write-Host "✅ Server is running at $ServerUrl" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "❌ Cannot connect to server at $ServerUrl" -ForegroundColor Red
        Write-Host "   Please start the server first: ./screenshot-server.exe --port 8080" -ForegroundColor Yellow
        return $false
    }
}

# Basic single window capture
function Invoke-BasicCapture {
    Write-Host "`n📸 Basic Window Capture Examples" -ForegroundColor Yellow
    
    $windows = @("Calculator", "Notepad", "explorer")
    
    foreach ($window in $windows) {
        try {
            $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
            $filename = "${window}_${timestamp}.png"
            $outputPath = Join-Path $OutputDir $filename
            
            $uri = "${ServerUrl}/api/screenshot?method=title&target=${window}&format=png"
            
            Write-Host "  Capturing $window..." -NoNewline
            Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
            
            if (Test-Path $outputPath) {
                $fileInfo = Get-Item $outputPath
                Write-Host " ✅ ($($fileInfo.Length) bytes)" -ForegroundColor Green
            }
        }
        catch {
            Write-Host " ❌ Failed" -ForegroundColor Red
        }
    }
}

# Desktop capture
function Invoke-DesktopCapture {
    Write-Host "`n🖥️  Desktop Capture Examples" -ForegroundColor Yellow
    
    # Full desktop
    try {
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $filename = "desktop_full_${timestamp}.png"
        $outputPath = Join-Path $OutputDir $filename
        
        $uri = "${ServerUrl}/api/screenshot?method=desktop&monitor=0&format=png"
        
        Write-Host "  Capturing full desktop..." -NoNewline
        Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
        
        $fileInfo = Get-Item $outputPath
        Write-Host " ✅ ($($fileInfo.Length) bytes)" -ForegroundColor Green
    }
    catch {
        Write-Host " ❌ Failed" -ForegroundColor Red
    }
    
    # Desktop region
    try {
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $filename = "desktop_region_${timestamp}.png"
        $outputPath = Join-Path $OutputDir $filename
        
        $uri = "${ServerUrl}/api/screenshot?method=desktop&region=100,100,800,600&format=png"
        
        Write-Host "  Capturing desktop region..." -NoNewline
        Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
        
        $fileInfo = Get-Item $outputPath
        Write-Host " ✅ ($($fileInfo.Length) bytes)" -ForegroundColor Green
    }
    catch {
        Write-Host " ❌ Failed" -ForegroundColor Red
    }
}

# Batch capture
function Invoke-BatchCapture {
    Write-Host "`n📦 Batch Capture Example" -ForegroundColor Yellow
    
    $batchConfig = @{
        targets = @(
            @{ method = "title"; target = "Calculator"; format = "png" }
            @{ method = "title"; target = "Notepad"; format = "jpeg"; quality = 85 }
            @{ method = "desktop"; monitor = 0; format = "png" }
        )
        options = @{
            parallel = $true
            timeout = 30
            fallback = $true
        }
    }
    
    try {
        Write-Host "  Executing batch capture..." -NoNewline
        
        $json = $batchConfig | ConvertTo-Json -Depth 5
        $response = Invoke-RestMethod -Uri "${ServerUrl}/api/screenshot/batch" -Method Post -Body $json -ContentType "application/json"
        
        Write-Host " ✅ Batch completed" -ForegroundColor Green
        
        foreach ($result in $response.results) {
            $status = if ($result.success) { "✅" } else { "❌" }
            Write-Host "    $status $($result.target) ($($result.size) bytes)"
        }
    }
    catch {
        Write-Host " ❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Hidden window capture
function Invoke-HiddenCapture {
    Write-Host "`n👻 Hidden Window Capture Examples" -ForegroundColor Yellow
    
    # Try to capture common hidden/minimized applications
    $targets = @(
        @{ target = "notepad.exe"; method = "process"; description = "Notepad process" }
        @{ target = "explorer.exe"; method = "tray"; description = "Explorer tray" }
        @{ target = "Calculator.exe"; method = "process"; description = "UWP Calculator" }
    )
    
    foreach ($item in $targets) {
        try {
            $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
            $filename = "hidden_$($item.target)_${timestamp}.png"
            $outputPath = Join-Path $OutputDir $filename
            
            $uri = "${ServerUrl}/api/screenshot?method=$($item.method)&target=$($item.target)&allow_hidden=true&format=png"
            
            Write-Host "  Capturing $($item.description)..." -NoNewline
            Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
            
            if (Test-Path $outputPath) {
                $fileInfo = Get-Item $outputPath
                Write-Host " ✅ ($($fileInfo.Length) bytes)" -ForegroundColor Green
            }
        }
        catch {
            Write-Host " ❌ Failed" -ForegroundColor Red
        }
    }
}

# Quality comparison
function Invoke-QualityComparison {
    Write-Host "`n🎨 Quality Comparison Examples" -ForegroundColor Yellow
    
    $qualities = @(50, 75, 90, 95)
    $target = "Calculator"
    
    foreach ($quality in $qualities) {
        try {
            $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
            $filename = "${target}_q${quality}_${timestamp}.jpg"
            $outputPath = Join-Path $OutputDir $filename
            
            $uri = "${ServerUrl}/api/screenshot?method=title&target=${target}&format=jpeg&quality=${quality}"
            
            Write-Host "  Quality ${quality}%..." -NoNewline
            Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
            
            if (Test-Path $outputPath) {
                $fileInfo = Get-Item $outputPath
                $sizeKB = [math]::Round($fileInfo.Length / 1024, 1)
                Write-Host " ✅ (${sizeKB} KB)" -ForegroundColor Green
            }
        }
        catch {
            Write-Host " ❌ Failed" -ForegroundColor Red
        }
    }
}

# Window discovery
function Invoke-WindowDiscovery {
    Write-Host "`n🔍 Window Discovery Examples" -ForegroundColor Yellow
    
    try {
        Write-Host "  Getting window list..." -NoNewline
        $response = Invoke-RestMethod -Uri "${ServerUrl}/api/windows" -Method Get
        Write-Host " ✅ Found $($response.windows.Count) windows" -ForegroundColor Green
        
        Write-Host "  Top 10 windows:" -ForegroundColor Cyan
        $response.windows | Select-Object -First 10 | ForEach-Object {
            $title = if ($_.title) { $_.title } else { "Untitled" }
            Write-Host "    • ${title} (Handle: $($_.handle))" -ForegroundColor Gray
        }
        
        # Save full list to JSON
        $jsonPath = Join-Path $OutputDir "windows_list.json"
        $response.windows | ConvertTo-Json -Depth 3 | Out-File -FilePath $jsonPath -Encoding UTF8
        Write-Host "    💾 Full list saved to: windows_list.json" -ForegroundColor Green
        
    }
    catch {
        Write-Host " ❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Chrome tab capture (if Chrome is running with debug port)
function Invoke-ChromeCapture {
    Write-Host "`n🌐 Chrome Tab Capture Examples" -ForegroundColor Yellow
    
    try {
        Write-Host "  Getting Chrome tabs..." -NoNewline
        $response = Invoke-RestMethod -Uri "${ServerUrl}/api/chrome/tabs" -Method Get -ErrorAction Stop
        Write-Host " ✅ Found $($response.tabs.Count) tabs" -ForegroundColor Green
        
        # Capture first few tabs
        $response.tabs | Select-Object -First 3 | ForEach-Object {
            try {
                $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
                $safeName = $_.title -replace '[^\w\-_\.]', '_'
                $filename = "chrome_${safeName}_${timestamp}.png"
                $outputPath = Join-Path $OutputDir $filename
                
                $uri = "${ServerUrl}/api/chrome/capture?tabId=$($_.id)&format=png"
                
                Write-Host "    Capturing '$($_.title)'..." -NoNewline
                Invoke-WebRequest -Uri $uri -OutFile $outputPath -ErrorAction Stop
                
                if (Test-Path $outputPath) {
                    $fileInfo = Get-Item $outputPath
                    Write-Host " ✅ ($($fileInfo.Length) bytes)" -ForegroundColor Green
                }
            }
            catch {
                Write-Host " ❌ Failed" -ForegroundColor Red
            }
        }
    }
    catch {
        Write-Host " ❌ Chrome not available (need --remote-debugging-port=9222)" -ForegroundColor Yellow
    }
}

# Performance test
function Invoke-PerformanceTest {
    Write-Host "`n⚡ Performance Test" -ForegroundColor Yellow
    
    $iterations = 10
    $target = "Calculator"
    $times = @()
    
    Write-Host "  Running $iterations iterations..." -ForegroundColor Cyan
    
    for ($i = 1; $i -le $iterations; $i++) {
        try {
            $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
            
            $uri = "${ServerUrl}/api/screenshot?method=title&target=${target}&format=png"
            $response = Invoke-WebRequest -Uri $uri -ErrorAction Stop
            
            $stopwatch.Stop()
            $times += $stopwatch.ElapsedMilliseconds
            
            Write-Host "    Iteration $i`: $($stopwatch.ElapsedMilliseconds)ms ($($response.RawContentLength) bytes)" -ForegroundColor Gray
        }
        catch {
            Write-Host "    Iteration $i`: Failed" -ForegroundColor Red
        }
    }
    
    if ($times.Count -gt 0) {
        $avgTime = ($times | Measure-Object -Average).Average
        $minTime = ($times | Measure-Object -Minimum).Minimum
        $maxTime = ($times | Measure-Object -Maximum).Maximum
        
        Write-Host "  📊 Results:" -ForegroundColor Cyan
        Write-Host "    Average: $([math]::Round($avgTime, 1))ms" -ForegroundColor Green
        Write-Host "    Min: ${minTime}ms" -ForegroundColor Green  
        Write-Host "    Max: ${maxTime}ms" -ForegroundColor Green
    }
}

# Main execution
if (Test-ServerConnection) {
    Write-Host "`n🚀 Running Screenshot Examples..." -ForegroundColor Cyan
    
    Invoke-BasicCapture
    Invoke-DesktopCapture
    Invoke-WindowDiscovery
    Invoke-QualityComparison
    Invoke-HiddenCapture
    Invoke-ChromeCapture
    Invoke-BatchCapture
    Invoke-PerformanceTest
    
    Write-Host "`n✨ All examples completed!" -ForegroundColor Green
    Write-Host "📁 Check the '$OutputDir' directory for generated screenshots." -ForegroundColor Cyan
    
    # Show directory contents
    $files = Get-ChildItem -Path $OutputDir -Filter "*.png", "*.jpg", "*.json" | Sort-Object LastWriteTime -Descending
    if ($files.Count -gt 0) {
        Write-Host "`n📸 Generated files ($($files.Count) total):" -ForegroundColor Yellow
        $files | Select-Object -First 10 | ForEach-Object {
            $sizeKB = [math]::Round($_.Length / 1024, 1)
            Write-Host "  • $($_.Name) (${sizeKB} KB)" -ForegroundColor Gray
        }
        if ($files.Count -gt 10) {
            Write-Host "  ... and $($files.Count - 10) more files" -ForegroundColor Gray
        }
    }
}

Write-Host "`n👋 Example script completed!" -ForegroundColor Cyan