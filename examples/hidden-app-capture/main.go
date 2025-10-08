package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/screenshot-mcp-server/internal/screenshot"
	"github.com/screenshot-mcp-server/pkg/types"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	// Create screenshot engine
	engine, err := screenshot.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create screenshot engine: %v", err)
	}

	switch command {
	case "discover-hidden":
		discoverHiddenWindows(engine)
	case "discover-tray":
		discoverTrayApps(engine)
	case "discover-cloaked":
		discoverCloakedWindows(engine)
	case "capture-hidden":
		captureHiddenApp(engine)
	case "capture-tray":
		captureTrayApp(engine)
	case "capture-pid":
		captureByPID(engine)
	case "test-fallbacks":
		testFallbackMethods(engine)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
	}
}

func showUsage() {
	fmt.Println("Hidden Application Screenshot Capture Demo")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Usage: go run main.go <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  discover-hidden    - Discover all hidden windows")
	fmt.Println("  discover-tray      - Discover system tray applications")
	fmt.Println("  discover-cloaked   - Discover DWM cloaked windows (UWP apps)")
	fmt.Println("  capture-hidden     - Capture screenshot of a hidden window by handle")
	fmt.Println("  capture-tray <app> - Capture screenshot of tray app by process name")
	fmt.Println("  capture-pid <pid>  - Capture any window from process ID")
	fmt.Println("  test-fallbacks     - Test all fallback capture methods")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go discover-hidden")
	fmt.Println("  go run main.go capture-tray notepad.exe")
	fmt.Println("  go run main.go capture-pid 1234")
	fmt.Println("  go run main.go test-fallbacks")
}

func discoverHiddenWindows(engine *screenshot.WindowsScreenshotEngine) {
	fmt.Println("Discovering hidden windows...")
	fmt.Println("============================")

	windows, err := engine.FindHiddenWindows()
	if err != nil {
		log.Fatalf("Failed to discover hidden windows: %v", err)
	}

	if len(windows) == 0 {
		fmt.Println("No hidden windows found.")
		return
	}

	fmt.Printf("Found %d hidden windows:\n\n", len(windows))
	
	for i, window := range windows {
		fmt.Printf("%d. Handle: 0x%X\n", i+1, window.Handle)
		fmt.Printf("   Title: %s\n", window.Title)
		fmt.Printf("   Class: %s\n", window.ClassName)
		fmt.Printf("   PID: %d\n", window.ProcessID)
		fmt.Printf("   State: %s\n", window.State)
		fmt.Printf("   Rect: %dx%d at (%d,%d)\n", 
			window.Rect.Width, window.Rect.Height, window.Rect.X, window.Rect.Y)
		fmt.Println()
	}
}

func discoverTrayApps(engine *screenshot.WindowsScreenshotEngine) {
	fmt.Println("Discovering system tray applications...")
	fmt.Println("=====================================")

	windows, err := engine.FindSystemTrayApps()
	if err != nil {
		log.Fatalf("Failed to discover tray apps: %v", err)
	}

	if len(windows) == 0 {
		fmt.Println("No system tray applications found.")
		return
	}

	fmt.Printf("Found %d system tray applications:\n\n", len(windows))
	
	// Group by process ID
	processWindows := make(map[uint32][]types.WindowInfo)
	for _, window := range windows {
		processWindows[window.ProcessID] = append(processWindows[window.ProcessID], window)
	}

	for pid, winList := range processWindows {
		fmt.Printf("Process ID %d:\n", pid)
		for _, window := range winList {
			fmt.Printf("  - Handle: 0x%X, Title: %s, Class: %s\n", 
				window.Handle, window.Title, window.ClassName)
		}
		fmt.Println()
	}
}

func discoverCloakedWindows(engine *screenshot.WindowsScreenshotEngine) {
	fmt.Println("Discovering DWM cloaked windows...")
	fmt.Println("==================================")

	windows, err := engine.FindCloakedWindows()
	if err != nil {
		log.Fatalf("Failed to discover cloaked windows: %v", err)
	}

	if len(windows) == 0 {
		fmt.Println("No cloaked windows found.")
		return
	}

	fmt.Printf("Found %d cloaked windows:\n\n", len(windows))
	
	for i, window := range windows {
		fmt.Printf("%d. Handle: 0x%X\n", i+1, window.Handle)
		fmt.Printf("   Title: %s\n", window.Title)
		fmt.Printf("   Class: %s\n", window.ClassName)
		fmt.Printf("   PID: %d\n", window.ProcessID)
		fmt.Printf("   State: %s\n", window.State)
		fmt.Printf("   Rect: %dx%d at (%d,%d)\n", 
			window.Rect.Width, window.Rect.Height, window.Rect.X, window.Rect.Y)
		fmt.Println()
	}
}

func captureHiddenApp(engine *screenshot.WindowsScreenshotEngine) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go capture-hidden <window_handle>")
		fmt.Println("Use 'discover-hidden' to find available handles")
		return
	}

	handleStr := os.Args[2]
	handle, err := strconv.ParseUint(strings.TrimPrefix(handleStr, "0x"), 16, 64)
	if err != nil {
		log.Fatalf("Invalid window handle: %v", err)
	}

	fmt.Printf("Capturing hidden window 0x%X with advanced fallbacks...\n", handle)
	fmt.Println("======================================================")

	// Create options for hidden window capture
	options := types.DefaultCaptureOptions()
	options.AllowHidden = true
	options.AllowMinimized = true
	options.AllowCloaked = true
	options.PreferredMethod = types.CaptureDWMThumbnail
	options.UseDWMThumbnails = true

	startTime := time.Now()
	buffer, err := engine.CaptureWithFallbacks(uintptr(handle), options)
	if err != nil {
		log.Fatalf("Failed to capture hidden window: %v", err)
	}
	captureTime := time.Since(startTime)

	fmt.Printf("✅ Successfully captured hidden window!\n")
	fmt.Printf("   Dimensions: %dx%d\n", buffer.Width, buffer.Height)
	fmt.Printf("   Format: %s\n", buffer.Format)
	fmt.Printf("   Size: %d bytes\n", len(buffer.Data))
	fmt.Printf("   Capture time: %v\n", captureTime)
	fmt.Printf("   Window title: %s\n", buffer.WindowInfo.Title)
	fmt.Printf("   Window class: %s\n", buffer.WindowInfo.ClassName)

	// Save as PNG
	saveScreenshot(buffer, "hidden_window_capture.png")
}

func captureTrayApp(engine *screenshot.WindowsScreenshotEngine) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go capture-tray <process_name>")
		fmt.Println("Example: go run main.go capture-tray notepad.exe")
		return
	}

	processName := os.Args[2]
	
	fmt.Printf("Capturing system tray application: %s\n", processName)
	fmt.Println("=====================================")

	startTime := time.Now()
	buffer, err := engine.CaptureTrayApp(processName, nil)
	if err != nil {
		log.Fatalf("Failed to capture tray app: %v", err)
	}
	captureTime := time.Since(startTime)

	fmt.Printf("✅ Successfully captured tray application!\n")
	fmt.Printf("   Process: %s\n", processName)
	fmt.Printf("   Dimensions: %dx%d\n", buffer.Width, buffer.Height)
	fmt.Printf("   Format: %s\n", buffer.Format)
	fmt.Printf("   Size: %d bytes\n", len(buffer.Data))
	fmt.Printf("   Capture time: %v\n", captureTime)
	fmt.Printf("   Window title: %s\n", buffer.WindowInfo.Title)
	fmt.Printf("   Window class: %s\n", buffer.WindowInfo.ClassName)

	// Save as PNG
	filename := fmt.Sprintf("tray_app_%s_capture.png", 
		strings.TrimSuffix(processName, ".exe"))
	saveScreenshot(buffer, filename)
}

func captureByPID(engine *screenshot.WindowsScreenshotEngine) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go capture-pid <process_id>")
		fmt.Println("Example: go run main.go capture-pid 1234")
		return
	}

	pidStr := os.Args[2]
	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		log.Fatalf("Invalid process ID: %v", err)
	}

	fmt.Printf("Capturing any window from process ID: %d\n", pid)
	fmt.Println("====================================")

	startTime := time.Now()
	buffer, err := engine.CaptureHiddenByPID(uint32(pid), nil)
	if err != nil {
		log.Fatalf("Failed to capture from PID: %v", err)
	}
	captureTime := time.Since(startTime)

	fmt.Printf("✅ Successfully captured from process!\n")
	fmt.Printf("   PID: %d\n", pid)
	fmt.Printf("   Dimensions: %dx%d\n", buffer.Width, buffer.Height)
	fmt.Printf("   Format: %s\n", buffer.Format)
	fmt.Printf("   Size: %d bytes\n", len(buffer.Data))
	fmt.Printf("   Capture time: %v\n", captureTime)
	fmt.Printf("   Window title: %s\n", buffer.WindowInfo.Title)
	fmt.Printf("   Window class: %s\n", buffer.WindowInfo.ClassName)
	fmt.Printf("   Window state: %s\n", buffer.WindowInfo.State)

	// Save as PNG
	filename := fmt.Sprintf("pid_%d_capture.png", pid)
	saveScreenshot(buffer, filename)
}

func testFallbackMethods(engine *screenshot.WindowsScreenshotEngine) {
	fmt.Println("Testing all capture fallback methods...")
	fmt.Println("=====================================")

	// First discover some windows to test with
	hidden, _ := engine.FindHiddenWindows()
	cloaked, _ := engine.FindCloakedWindows()
	
	testWindows := make([]types.WindowInfo, 0)
	testWindows = append(testWindows, hidden...)
	testWindows = append(testWindows, cloaked...)

	if len(testWindows) == 0 {
		fmt.Println("No hidden or cloaked windows found to test with.")
		return
	}

	// Test each capture method
	methods := []types.CaptureMethod{
		types.CaptureDWMThumbnail,
		types.CapturePrintWindow,
		types.CaptureWMPrint,
		types.CaptureStealthRestore,
	}

	fmt.Printf("Testing with %d windows and %d methods...\n\n", 
		len(testWindows), len(methods))

	results := make(map[types.CaptureMethod]int)
	total := 0

	for _, window := range testWindows {
		fmt.Printf("Testing window: %s (0x%X)\n", window.Title, window.Handle)
		
		for _, method := range methods {
			options := types.DefaultCaptureOptions()
			options.PreferredMethod = method
			options.FallbackMethods = []types.CaptureMethod{} // No fallbacks for pure testing
			options.AllowHidden = true
			options.AllowCloaked = true

			startTime := time.Now()
			buffer, err := engine.CaptureWithFallbacks(window.Handle, options)
			duration := time.Since(startTime)

			if err == nil {
				fmt.Printf("  ✅ %s: Success (%v, %dx%d)\n", 
					method, duration, buffer.Width, buffer.Height)
				results[method]++
			} else {
				fmt.Printf("  ❌ %s: Failed (%v)\n", method, err)
			}
			total++
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("=== FALLBACK METHOD RESULTS ===")
	for _, method := range methods {
		success := results[method]
		successRate := float64(success) / float64(len(testWindows)) * 100
		fmt.Printf("%s: %d/%d (%.1f%% success rate)\n", 
			method, success, len(testWindows), successRate)
	}
	
	totalTests := len(testWindows) * len(methods)
	totalSuccess := 0
	for _, count := range results {
		totalSuccess += count
	}
	
	fmt.Printf("\nOverall: %d/%d tests passed (%.1f%% success rate)\n", 
		totalSuccess, totalTests, float64(totalSuccess)/float64(totalTests)*100)
}

func saveScreenshot(buffer *types.ScreenshotBuffer, filename string) {
	// For this example, we'll save as a simple JSON metadata file
	// In a real implementation, you'd encode to PNG/JPEG
	
	metadata := map[string]interface{}{
		"width":      buffer.Width,
		"height":     buffer.Height,
		"format":     buffer.Format,
		"timestamp":  buffer.Timestamp,
		"window":     buffer.WindowInfo,
		"data_size":  len(buffer.Data),
		"data_base64": base64.StdEncoding.EncodeToString(buffer.Data[:min(1024, len(buffer.Data))]), // First 1KB as sample
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal metadata: %v", err)
		return
	}

	metadataFile := strings.TrimSuffix(filename, ".png") + "_metadata.json"
	err = os.WriteFile(metadataFile, data, 0644)
	if err != nil {
		log.Printf("Failed to save metadata: %v", err)
		return
	}

	fmt.Printf("   💾 Saved metadata to: %s\n", metadataFile)
	fmt.Printf("   📸 Raw image data: %d bytes (%s format)\n", 
		len(buffer.Data), buffer.Format)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}