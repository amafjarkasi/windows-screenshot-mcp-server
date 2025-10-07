package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/screenshot-mcp-server/internal/chrome"
	"github.com/screenshot-mcp-server/internal/screenshot"
	"github.com/screenshot-mcp-server/pkg/types"
)

var (
	serverURL string
	format    string
	quality   int
	output    string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "mcpctl",
	Short: "CLI tool for Screenshot MCP Server",
	Long: `mcpctl is a command line interface for the Screenshot MCP Server.
It allows you to take screenshots, manage windows, and interact with Chrome tabs.`,
}

// screenshotCmd represents the screenshot command
var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Take screenshots",
	Long:  `Take screenshots of windows by various methods.`,
}

// windowsCmd represents the windows command
var windowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Manage windows",
	Long:  `List and manage windows.`,
}

// chromeCmd represents the chrome command
var chromeCmd = &cobra.Command{
	Use:   "chrome",
	Short: "Chrome integration",
	Long:  `Interact with Chrome browser instances and tabs.`,
}

// Screenshot commands
var captureByTitleCmd = &cobra.Command{
	Use:   "title [window-title]",
	Short: "Capture screenshot by window title",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		captureScreenshot("title", args[0])
	},
}

var captureByPIDCmd = &cobra.Command{
	Use:   "pid [process-id]",
	Short: "Capture screenshot by process ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		captureScreenshot("pid", args[0])
	},
}

var captureByClassCmd = &cobra.Command{
	Use:   "class [class-name]",
	Short: "Capture screenshot by window class",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		captureScreenshot("class", args[0])
	},
}

// Window commands
var listWindowsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all windows",
	Run: func(cmd *cobra.Command, args []string) {
		listWindows()
	},
}

// Chrome commands
var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List Chrome instances",
	Run: func(cmd *cobra.Command, args []string) {
		listChromeInstances()
	},
}

var listTabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List Chrome tabs",
	Run: func(cmd *cobra.Command, args []string) {
		listChromeTabs()
	},
}

var captureTabCmd = &cobra.Command{
	Use:   "capture [tab-id]",
	Short: "Capture Chrome tab screenshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		captureChromeTab(args[0])
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Screenshot server URL")
	rootCmd.PersistentFlags().StringVar(&format, "format", "png", "Image format (png, jpeg)")
	rootCmd.PersistentFlags().IntVar(&quality, "quality", 95, "Image quality (1-100)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "", "Output file path")

	// Add commands
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(windowsCmd)
	rootCmd.AddCommand(chromeCmd)

	// Screenshot subcommands
	screenshotCmd.AddCommand(captureByTitleCmd)
	screenshotCmd.AddCommand(captureByPIDCmd)
	screenshotCmd.AddCommand(captureByClassCmd)

	// Windows subcommands
	windowsCmd.AddCommand(listWindowsCmd)

	// Chrome subcommands
	chromeCmd.AddCommand(listInstancesCmd)
	chromeCmd.AddCommand(listTabsCmd)
	chromeCmd.AddCommand(captureTabCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Local screenshot functions (bypass server for testing)
func captureScreenshot(method, target string) {
	fmt.Printf("Capturing screenshot: method=%s, target=%s\n", method, target)

	// Initialize screenshot engine directly
	engine, err := screenshot.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create screenshot engine: %v", err)
	}

	options := types.DefaultCaptureOptions()
	var buffer *types.ScreenshotBuffer

	switch method {
	case "title":
		buffer, err = engine.CaptureByTitle(target, options)
	case "pid":
		// Parse PID (simplified for demo)
		log.Fatalf("PID capture not implemented in CLI demo")
	case "class":
		buffer, err = engine.CaptureByClassName(target, options)
	default:
		log.Fatalf("Unknown method: %s", method)
	}

	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	fmt.Printf("Screenshot captured: %dx%d, %d bytes, DPI: %d\n",
		buffer.Width, buffer.Height, len(buffer.Data), buffer.DPI)

	// Save to file if requested
	if output != "" {
		// This would require implementing image encoding
		fmt.Printf("Saving to %s (not implemented in demo)\n", output)
	}

	// Show window info
	fmt.Printf("Window: %s (PID: %d, Class: %s)\n",
		buffer.WindowInfo.Title,
		buffer.WindowInfo.ProcessID,
		buffer.WindowInfo.ClassName)
}

func listWindows() {
	fmt.Println("Listing windows (not implemented in CLI demo)")
	// This would require implementing window enumeration
}

func listChromeInstances() {
	fmt.Println("Discovering Chrome instances...")

	manager := chrome.NewManager()
	instances, err := manager.DiscoverInstances()
	if err != nil {
		log.Fatalf("Failed to discover Chrome instances: %v", err)
	}

	if len(instances) == 0 {
		fmt.Println("No Chrome instances found")
		return
	}

	fmt.Printf("Found %d Chrome instance(s):\n", len(instances))
	for i, instance := range instances {
		fmt.Printf("  [%d] PID: %d, Port: %d, Version: %s\n",
			i+1, instance.PID, instance.DebugPort, instance.Version)
	}
}

func listChromeTabs() {
	fmt.Println("Discovering Chrome tabs...")

	manager := chrome.NewManager()
	instances, err := manager.DiscoverInstances()
	if err != nil {
		log.Fatalf("Failed to discover Chrome instances: %v", err)
	}

	if len(instances) == 0 {
		fmt.Println("No Chrome instances found")
		return
	}

	totalTabs := 0
	for _, instance := range instances {
		fmt.Printf("\nChrome instance (PID: %d, Port: %d):\n", instance.PID, instance.DebugPort)

		tabs, err := manager.GetTabs(&instance)
		if err != nil {
			fmt.Printf("  Error getting tabs: %v\n", err)
			continue
		}

		for i, tab := range tabs {
			fmt.Printf("  [%d] %s\n", i+1, tab.Title)
			fmt.Printf("      ID: %s\n", tab.ID)
			fmt.Printf("      URL: %s\n", tab.URL)
			if tab.Active {
				fmt.Printf("      (Active)\n")
			}
			fmt.Println()
		}

		totalTabs += len(tabs)
	}

	fmt.Printf("Total tabs found: %d\n", totalTabs)
}

func captureChromeTab(tabID string) {
	fmt.Printf("Capturing Chrome tab: %s\n", tabID)

	manager := chrome.NewManager()
	instances, err := manager.DiscoverInstances()
	if err != nil {
		log.Fatalf("Failed to discover Chrome instances: %v", err)
	}

	// Find the tab
	var targetTab *types.ChromeTab
	for _, instance := range instances {
		tabs, err := manager.GetTabs(&instance)
		if err != nil {
			continue
		}

		for _, tab := range tabs {
			if tab.ID == tabID {
				targetTab = &tab
				break
			}
		}
		if targetTab != nil {
			break
		}
	}

	if targetTab == nil {
		log.Fatalf("Tab not found: %s", tabID)
	}

	fmt.Printf("Found tab: %s\n", targetTab.Title)

	// Capture screenshot
	options := types.DefaultCaptureOptions()
	buffer, err := manager.CaptureTab(targetTab, options)
	if err != nil {
		log.Fatalf("Failed to capture tab screenshot: %v", err)
	}

	fmt.Printf("Screenshot captured: %dx%d, %d bytes\n",
		buffer.Width, buffer.Height, len(buffer.Data))

	// Save to file if requested
	if output != "" {
		fmt.Printf("Saving to %s (not implemented in demo)\n", output)
	}
}

// Utility function to pretty print JSON
func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(data))
}