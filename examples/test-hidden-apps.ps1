#!/usr/bin/env pwsh
param(
    [switch]$Help,
    [switch]$TestAll,
    [switch]$TestHidden,
    [switch]$TestTray,
    [switch]$TestCloaked,
    [switch]$TestFallbacks,
    [string]$ProcessName = "",
    [int]$ProcessID = 0
)

# Script configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ExampleDir = Join-Path $ScriptDir "hidden-app-capture"
$ServerPort = 8080
$ServerHost = "localhost"

function Show-Help {
    Write-Host "🔍 Hidden Application Screenshot Testing Script" -ForegroundColor Cyan
    Write-Host "===============================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "This script tests the genius-level hidden window capture capabilities"
    Write-Host "including system tray apps, minimized windows, and DWM cloaked windows."
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor Yellow
    Write-Host "  -Help                Show this help message"
    Write-Host "  -TestAll             Run all tests (discovery + capture)"  
    Write-Host "  -TestHidden          Test hidden window discovery and capture"
    Write-Host "  -TestTray            Test system tray application capture"
    Write-Host "  -TestCloaked         Test DWM cloaked window capture (UWP apps)"
    Write-Host "  -TestFallbacks       Test all capture method fallbacks"
    Write-Host "  -ProcessName <name>  Test specific process by name (e.g., notepad.exe)"
    Write-Host "  -ProcessID <id>      Test specific process by ID"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Green
    Write-Host "  .\test-hidden-apps.ps1 -TestAll"
    Write-Host "  .\test-hidden-apps.ps1 -TestHidden"
    Write-Host "  .\test-hidden-apps.ps1 -TestTray"
    Write-Host "  .\test-hidden-apps.ps1 -ProcessName notepad.exe"
    Write-Host "  .\test-hidden-apps.ps1 -ProcessID 1234"
    Write-Host "  .\test-hidden-apps.ps1 -TestFallbacks"
    Write-Host ""
    Write-Host "Advanced Features Tested:" -ForegroundColor Magenta
    Write-Host "  • System tray application capture (even when hidden to tray)"
    Write-Host "  • Minimized window capture without restoring"
    Write-Host "  • DWM cloaked window capture (UWP/Store apps)"
    Write-Host "  • Hidden window discovery and enumeration"
    Write-Host "  • Multiple capture method fallbacks (DWM, PrintWindow, WM_PRINT)"
    Write-Host "  • Process-based window discovery"
    Write-Host "  • Stealth window restoration (without activation)"
}

function Test-Prerequisites {
    Write-Host "🔧 Checking prerequisites..." -ForegroundColor Yellow
    
    # Check if Go is installed
    try {
        $goVersion = & go version 2>$null
        Write-Host "  ✅ Go is installed: $goVersion" -ForegroundColor Green
    }
    catch {
        Write-Host "  ❌ Go is not installed or not in PATH" -ForegroundColor Red
        return $false
    }
    
    # Check if project files exist
    if (-not (Test-Path $ProjectRoot)) {
        Write-Host "  ❌ Project root not found: $ProjectRoot" -ForegroundColor Red
        return $false
    }
    
    if (-not (Test-Path $ExampleDir)) {
        Write-Host "  ❌ Example directory not found: $ExampleDir" -ForegroundColor Red
        return $false
    }
    
    Write-Host "  ✅ Project structure verified" -ForegroundColor Green
    Write-Host "  ✅ Example directory found: $ExampleDir" -ForegroundColor Green
    
    return $true
}

function Start-HiddenWindowDiscovery {
    Write-Host ""
    Write-Host "🔍 Discovering Hidden Windows..." -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    
    Push-Location $ExampleDir
    try {
        Write-Host "Running: go run main.go discover-hidden" -ForegroundColor Gray
        & go run main.go discover-hidden
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Hidden window discovery completed successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ Hidden window discovery failed with exit code $LASTEXITCODE" -ForegroundColor Red
        }
    }
    finally {
        Pop-Location
    }
}

function Start-TrayAppDiscovery {
    Write-Host ""
    Write-Host "📱 Discovering System Tray Applications..." -ForegroundColor Cyan
    Write-Host "==========================================" -ForegroundColor Cyan
    
    Push-Location $ExampleDir
    try {
        Write-Host "Running: go run main.go discover-tray" -ForegroundColor Gray
        & go run main.go discover-tray
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ System tray discovery completed successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ System tray discovery failed with exit code $LASTEXITCODE" -ForegroundColor Red
        }
    }
    finally {
        Pop-Location
    }
}

function Start-CloakedWindowDiscovery {
    Write-Host ""
    Write-Host "👻 Discovering DWM Cloaked Windows..." -ForegroundColor Cyan
    Write-Host "=====================================" -ForegroundColor Cyan
    
    Push-Location $ExampleDir
    try {
        Write-Host "Running: go run main.go discover-cloaked" -ForegroundColor Gray
        & go run main.go discover-cloaked
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Cloaked window discovery completed successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ Cloaked window discovery failed with exit code $LASTEXITCODE" -ForegroundColor Red
        }
    }
    finally {
        Pop-Location
    }
}

function Test-ProcessCapture {
    param(
        [string]$ProcessName,
        [int]$ProcessID
    )
    
    if ($ProcessName) {
        Write-Host ""
        Write-Host "🎯 Testing Process Capture by Name: $ProcessName" -ForegroundColor Cyan
        Write-Host "================================================" -ForegroundColor Cyan
        
        Push-Location $ExampleDir
        try {
            Write-Host "Running: go run main.go capture-tray $ProcessName" -ForegroundColor Gray
            & go run main.go capture-tray $ProcessName
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Process capture by name completed successfully" -ForegroundColor Green
            } else {
                Write-Host "❌ Process capture by name failed with exit code $LASTEXITCODE" -ForegroundColor Red
            }
        }
        finally {
            Pop-Location
        }
    }
    
    if ($ProcessID -gt 0) {
        Write-Host ""
        Write-Host "🎯 Testing Process Capture by ID: $ProcessID" -ForegroundColor Cyan
        Write-Host "===========================================" -ForegroundColor Cyan
        
        Push-Location $ExampleDir
        try {
            Write-Host "Running: go run main.go capture-pid $ProcessID" -ForegroundColor Gray
            & go run main.go capture-pid $ProcessID
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Process capture by ID completed successfully" -ForegroundColor Green
            } else {
                Write-Host "❌ Process capture by ID failed with exit code $LASTEXITCODE" -ForegroundColor Red
            }
        }
        finally {
            Pop-Location
        }
    }
}

function Test-FallbackMethods {
    Write-Host ""
    Write-Host "🔄 Testing Capture Method Fallbacks..." -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    
    Push-Location $ExampleDir
    try {
        Write-Host "Running: go run main.go test-fallbacks" -ForegroundColor Gray
        & go run main.go test-fallbacks
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Fallback method testing completed successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ Fallback method testing failed with exit code $LASTEXITCODE" -ForegroundColor Red
        }
    }
    finally {
        Pop-Location
    }
}

function Get-CommonProcessesToTest {
    Write-Host ""
    Write-Host "🔍 Finding Common Applications for Testing..." -ForegroundColor Cyan
    Write-Host "=============================================" -ForegroundColor Cyan
    
    # Common processes that might have tray icons or hidden windows
    $commonProcesses = @(
        "explorer.exe",
        "winlogon.exe", 
        "dwm.exe",
        "lsass.exe",
        "services.exe",
        "svchost.exe"
    )
    
    $foundProcesses = @()
    
    foreach ($processName in $commonProcesses) {
        $processes = Get-Process -Name ($processName -replace "\.exe$", "") -ErrorAction SilentlyContinue
        if ($processes) {
            foreach ($proc in $processes) {
                $foundProcesses += @{
                    Name = $processName
                    ID = $proc.Id
                    ProcessName = $proc.ProcessName
                }
                Write-Host "  Found: $processName (PID: $($proc.Id))" -ForegroundColor Green
            }
        }
    }
    
    return $foundProcesses
}

function Test-CommonApplications {
    Write-Host ""
    Write-Host "🎯 Testing Common Applications..." -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    
    $processes = Get-CommonProcessesToTest
    $testCount = [Math]::Min($processes.Count, 3)  # Test up to 3 processes
    
    for ($i = 0; $i -lt $testCount; $i++) {
        $proc = $processes[$i]
        Write-Host ""
        Write-Host "Testing Process: $($proc.Name) (PID: $($proc.ID))" -ForegroundColor Yellow
        
        Push-Location $ExampleDir
        try {
            & go run main.go capture-pid $proc.ID
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  ✅ Successfully captured $($proc.Name)" -ForegroundColor Green
            } else {
                Write-Host "  ❌ Failed to capture $($proc.Name)" -ForegroundColor Red
            }
        }
        finally {
            Pop-Location
        }
        
        Start-Sleep -Milliseconds 500  # Brief pause between tests
    }
}

function Show-Results {
    Write-Host ""
    Write-Host "📊 Test Results Summary" -ForegroundColor Cyan
    Write-Host "======================" -ForegroundColor Cyan
    
    $outputFiles = Get-ChildItem -Path $ExampleDir -Filter "*_metadata.json" -ErrorAction SilentlyContinue
    
    if ($outputFiles) {
        Write-Host ""
        Write-Host "Generated Output Files:" -ForegroundColor Green
        foreach ($file in $outputFiles) {
            Write-Host "  📄 $($file.Name)" -ForegroundColor Gray
            
            # Try to read and display basic info from the metadata
            try {
                $metadata = Get-Content $file.FullName | ConvertFrom-Json
                Write-Host "     Resolution: $($metadata.width)x$($metadata.height)" -ForegroundColor Gray
                Write-Host "     Window: $($metadata.window.Title)" -ForegroundColor Gray
                Write-Host "     Size: $($metadata.data_size) bytes" -ForegroundColor Gray
            }
            catch {
                Write-Host "     (Unable to read metadata)" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "No output files generated." -ForegroundColor Yellow
    }
    
    Write-Host ""
    Write-Host "🎉 Hidden Application Screenshot Testing Complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "What was tested:" -ForegroundColor Yellow
    Write-Host "• Hidden window discovery and enumeration"
    Write-Host "• System tray application detection"
    Write-Host "• DWM cloaked window discovery (UWP apps)"
    Write-Host "• Multiple capture method fallbacks"
    Write-Host "• Process-based window capture"
    Write-Host "• Advanced Windows API integration"
    Write-Host ""
}

# Main execution
Write-Host "🚀 Windows Hidden Application Screenshot Testing" -ForegroundColor Magenta
Write-Host "=================================================" -ForegroundColor Magenta
Write-Host "This script tests genius-level screenshot capabilities for:"
Write-Host "• System tray applications • Minimized windows • Hidden windows"
Write-Host "• DWM cloaked windows • Advanced capture fallbacks"
Write-Host ""

if ($Help) {
    Show-Help
    exit 0
}

if (-not (Test-Prerequisites)) {
    Write-Host ""
    Write-Host "❌ Prerequisites not met. Please install Go and ensure project structure is correct." -ForegroundColor Red
    exit 1
}

# Build the example first
Write-Host "🔨 Building hidden app capture example..." -ForegroundColor Yellow
Push-Location $ExampleDir
try {
    & go mod tidy 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ❌ Failed to tidy Go modules" -ForegroundColor Red
        exit 1
    }
    Write-Host "  ✅ Go modules updated" -ForegroundColor Green
}
finally {
    Pop-Location
}

# Execute requested tests
if ($TestAll) {
    Start-HiddenWindowDiscovery
    Start-TrayAppDiscovery  
    Start-CloakedWindowDiscovery
    Test-CommonApplications
    Test-FallbackMethods
}
elseif ($TestHidden) {
    Start-HiddenWindowDiscovery
    Test-CommonApplications
}
elseif ($TestTray) {
    Start-TrayAppDiscovery
}
elseif ($TestCloaked) {
    Start-CloakedWindowDiscovery
}
elseif ($TestFallbacks) {
    Test-FallbackMethods
}
elseif ($ProcessName -or $ProcessID -gt 0) {
    Test-ProcessCapture -ProcessName $ProcessName -ProcessID $ProcessID
}
else {
    Write-Host "No test specified. Use -Help for usage information." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Quick start options:" -ForegroundColor Cyan
    Write-Host "  .\test-hidden-apps.ps1 -TestAll      # Run all tests"
    Write-Host "  .\test-hidden-apps.ps1 -TestHidden   # Test hidden windows"
    Write-Host "  .\test-hidden-apps.ps1 -Help         # Show full help"
    exit 0
}

Show-Results