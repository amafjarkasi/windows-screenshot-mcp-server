package types

import (
	"image"
	"time"
)

// ScreenshotRequest represents a request to capture a screenshot
type ScreenshotRequest struct {
	Method        string            `json:"method"`         // "title", "pid", "handle", "class"
	Target        string            `json:"target"`         // Window title, PID, handle, or class name
	Format        ImageFormat       `json:"format"`         // Output format
	Quality       int               `json:"quality"`        // JPEG quality (1-100)
	IncludeCursor bool              `json:"include_cursor"` // Include mouse cursor
	Region        *Rectangle        `json:"region"`         // Specific region to capture
	Options       map[string]string `json:"options"`        // Additional options
}

// ScreenshotResponse represents the response containing screenshot data
type ScreenshotResponse struct {
	Success   bool      `json:"success"`
	Data      string    `json:"data"`       // Base64 encoded image data
	Format    string    `json:"format"`     // Actual format used
	Width     int       `json:"width"`      // Image width
	Height    int       `json:"height"`     // Image height
	Size      int64     `json:"size"`       // Size in bytes
	Timestamp time.Time `json:"timestamp"`  // When captured
	Metadata  Metadata  `json:"metadata"`   // Additional metadata
	Error     string    `json:"error"`      // Error message if failed
}

// WindowInfo contains information about a window
type WindowInfo struct {
	Handle     uintptr   `json:"handle"`      // Windows HWND
	Title      string    `json:"title"`       // Window title
	ClassName  string    `json:"class_name"`  // Window class name
	ProcessID  uint32    `json:"process_id"`  // Process ID
	ThreadID   uint32    `json:"thread_id"`   // Thread ID
	Rect       Rectangle `json:"rect"`        // Window rectangle
	ClientRect Rectangle `json:"client_rect"` // Client area rectangle
	State      string    `json:"state"`       // "visible", "minimized", "maximized", "hidden"
	ZOrder     int       `json:"z_order"`     // Z-order position
	IsVisible  bool      `json:"is_visible"`  // Whether window is visible
	IsTopMost  bool      `json:"is_topmost"`  // Whether window is always on top
	Monitor    int       `json:"monitor"`     // Monitor index
}

// ChromeTab represents a Chrome browser tab
type ChromeTab struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	Description string `json:"description"`
	WindowID    int    `json:"windowId"`
	DevToolsURL string `json:"devtoolsFrontendUrl"`
	WebSocketURL string `json:"webSocketDebuggerUrl"`
	Active      bool   `json:"active"`
}

// ChromeInstance represents a Chrome browser instance
type ChromeInstance struct {
	PID         uint32      `json:"pid"`
	DebugPort   int         `json:"debug_port"`
	ProfilePath string      `json:"profile_path"`
	Tabs        []ChromeTab `json:"tabs"`
	Version     string      `json:"version"`
	UserAgent   string      `json:"user_agent"`
}

// Rectangle represents a rectangular area
type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Point represents a 2D point
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Size represents dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// MonitorInfo contains information about a display monitor
type MonitorInfo struct {
	Index     int       `json:"index"`
	Primary   bool      `json:"primary"`
	Rect      Rectangle `json:"rect"`
	WorkArea  Rectangle `json:"work_area"`
	DPI       int       `json:"dpi"`
	ScaleFactor float64 `json:"scale_factor"`
	Name      string    `json:"name"`
}

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatPNG  ImageFormat = "png"
	FormatJPEG ImageFormat = "jpeg"
	FormatBMP  ImageFormat = "bmp"
	FormatWebP ImageFormat = "webp"
)

// ScreenshotBuffer contains raw image data with metadata
type ScreenshotBuffer struct {
	Data        []byte     `json:"-"`      // Raw image data (BGRA)
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	Stride      int        `json:"stride"`  // Bytes per row
	Format      string     `json:"format"`  // "BGRA32"
	DPI         int        `json:"dpi"`
	Timestamp   time.Time  `json:"timestamp"`
	SourceRect  Rectangle  `json:"source_rect"`
	WindowInfo  WindowInfo `json:"window_info"`
	MonitorInfo MonitorInfo `json:"monitor_info"`
}

// Metadata contains additional information about a screenshot
type Metadata struct {
	CaptureMethod   string            `json:"capture_method"`   // How it was captured
	ProcessingTime  time.Duration     `json:"processing_time"`  // Time to process
	WindowVisible   bool              `json:"window_visible"`   // Was window visible
	WindowMinimized bool              `json:"window_minimized"` // Was window minimized
	DPIScaling      float64           `json:"dpi_scaling"`      // DPI scale factor
	ColorDepth      int               `json:"color_depth"`      // Bits per pixel
	Properties      map[string]string `json:"properties"`       // Additional properties
}

// StreamSession represents an active streaming session
type StreamSession struct {
	ID        string      `json:"id"`
	WindowID  uintptr     `json:"window_id"`
	FPS       int         `json:"fps"`
	Quality   int         `json:"quality"`
	Format    ImageFormat `json:"format"`
	Active    bool        `json:"active"`
	StartTime time.Time   `json:"start_time"`
	FrameCount int64      `json:"frame_count"`
	BytesSent  int64      `json:"bytes_sent"`
}

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Interfaces

// ScreenshotEngine defines the core screenshot functionality
type ScreenshotEngine interface {
	// Capture by window handle
	CaptureByHandle(handle uintptr, options *CaptureOptions) (*ScreenshotBuffer, error)
	
	// Capture by window title (supports partial matches)
	CaptureByTitle(title string, options *CaptureOptions) (*ScreenshotBuffer, error)
	
	// Capture by process ID
	CaptureByPID(pid uint32, options *CaptureOptions) (*ScreenshotBuffer, error)
	
	// Capture by window class name
	CaptureByClassName(className string, options *CaptureOptions) (*ScreenshotBuffer, error)
	
	// Capture full screen
	CaptureFullScreen(monitor int, options *CaptureOptions) (*ScreenshotBuffer, error)
}

// WindowManager defines window management operations
type WindowManager interface {
	// Enumerate all windows with optional filtering
	EnumerateWindows(filter *WindowFilter) ([]WindowInfo, error)
	
	// Get window information by handle
	GetWindowInfo(handle uintptr) (*WindowInfo, error)
	
	// Set window position and size
	SetWindowPos(handle uintptr, rect Rectangle) error
	
	// Show/hide window
	SetWindowVisible(handle uintptr, visible bool) error
	
	// Minimize/restore window
	SetWindowState(handle uintptr, state string) error
	
	// Bring window to foreground
	BringToForeground(handle uintptr) error
}

// ChromeManager defines Chrome browser interaction
type ChromeManager interface {
	// Discover Chrome instances
	DiscoverInstances() ([]ChromeInstance, error)
	
	// Get tabs for a specific Chrome instance
	GetTabs(instance *ChromeInstance) ([]ChromeTab, error)
	
	// Capture screenshot of a tab
	CaptureTab(tab *ChromeTab, options *CaptureOptions) (*ScreenshotBuffer, error)
	
	// Execute JavaScript in tab context
	ExecuteScript(tab *ChromeTab, script string) (interface{}, error)
}

// ImageProcessor defines image processing operations
type ImageProcessor interface {
	// Encode buffer to specific format
	Encode(buffer *ScreenshotBuffer, format ImageFormat, quality int) ([]byte, error)
	
	// Decode image data to buffer
	Decode(data []byte) (*ScreenshotBuffer, error)
	
	// Resize image
	Resize(buffer *ScreenshotBuffer, width, height int) (*ScreenshotBuffer, error)
	
	// Crop image
	Crop(buffer *ScreenshotBuffer, rect Rectangle) (*ScreenshotBuffer, error)
	
	// Convert to Go image.Image
	ToImage(buffer *ScreenshotBuffer) (image.Image, error)
}

// StreamManager defines streaming functionality
type StreamManager interface {
	// Start streaming session
	StartSession(windowID uintptr, options *StreamOptions) (*StreamSession, error)
	
	// Stop streaming session
	StopSession(sessionID string) error
	
	// Get active sessions
	GetActiveSessions() ([]*StreamSession, error)
	
	// Update session parameters
	UpdateSession(sessionID string, options *StreamOptions) error
}

// Options structs

// CaptureOptions defines options for screenshot capture
type CaptureOptions struct {
	IncludeCursor    bool       `json:"include_cursor"`
	IncludeFrame     bool       `json:"include_frame"`
	Region           *Rectangle `json:"region"`
	ScaleFactor      float64    `json:"scale_factor"`
	AllowMinimized   bool       `json:"allow_minimized"`
	RestoreWindow    bool       `json:"restore_window"`
	WaitForVisible   time.Duration `json:"wait_for_visible"`
	RetryCount       int        `json:"retry_count"`
	CustomProperties map[string]string `json:"custom_properties"`
}

// WindowFilter defines filtering options for window enumeration
type WindowFilter struct {
	TitleContains  string   `json:"title_contains"`
	ClassNames     []string `json:"class_names"`
	ProcessIDs     []uint32 `json:"process_ids"`
	VisibleOnly    bool     `json:"visible_only"`
	MinimumSize    *Size    `json:"minimum_size"`
	MaximumSize    *Size    `json:"maximum_size"`
	ExcludeSystem  bool     `json:"exclude_system"`
}

// StreamOptions defines options for streaming
type StreamOptions struct {
	FPS            int         `json:"fps"`
	Quality        int         `json:"quality"`
	Format         ImageFormat `json:"format"`
	MaxWidth       int         `json:"max_width"`
	MaxHeight      int         `json:"max_height"`
	BufferSize     int         `json:"buffer_size"`
	CompressionLevel int       `json:"compression_level"`
}

// DefaultCaptureOptions returns sensible defaults for screenshot capture
func DefaultCaptureOptions() *CaptureOptions {
	return &CaptureOptions{
		IncludeCursor:    false,
		IncludeFrame:     true,
		ScaleFactor:      1.0,
		AllowMinimized:   true,
		RestoreWindow:    false,
		WaitForVisible:   time.Second * 2,
		RetryCount:       3,
		CustomProperties: make(map[string]string),
	}
}

// DefaultStreamOptions returns sensible defaults for streaming
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		FPS:              10,
		Quality:          75,
		Format:           FormatJPEG,
		MaxWidth:         1920,
		MaxHeight:        1080,
		BufferSize:       5,
		CompressionLevel: 6,
	}
}

// ToRect converts Rectangle to image.Rectangle
func (r Rectangle) ToRect() image.Rectangle {
	return image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)
}

// FromRect converts image.Rectangle to Rectangle
func FromRect(r image.Rectangle) Rectangle {
	return Rectangle{
		X:      r.Min.X,
		Y:      r.Min.Y,
		Width:  r.Dx(),
		Height: r.Dy(),
	}
}

// Contains checks if a point is within the rectangle
func (r Rectangle) Contains(p Point) bool {
	return p.X >= r.X && p.X < r.X+r.Width &&
		p.Y >= r.Y && p.Y < r.Y+r.Height
}

// Intersect returns the intersection of two rectangles
func (r Rectangle) Intersect(other Rectangle) Rectangle {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)
	
	if x2 <= x1 || y2 <= y1 {
		return Rectangle{} // No intersection
	}
	
	return Rectangle{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

// Union returns the union of two rectangles
func (r Rectangle) Union(other Rectangle) Rectangle {
	x1 := min(r.X, other.X)
	y1 := min(r.Y, other.Y)
	x2 := max(r.X+r.Width, other.X+other.Width)
	y2 := max(r.Y+r.Height, other.Y+other.Height)
	
	return Rectangle{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}