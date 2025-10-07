package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/screenshot-mcp-server/internal/chrome"
	"github.com/screenshot-mcp-server/internal/screenshot"
	"github.com/screenshot-mcp-server/internal/ws"
	"github.com/screenshot-mcp-server/pkg/types"
	"go.uber.org/zap"
)

// Server represents the MCP screenshot server
type Server struct {
	engine         types.ScreenshotEngine
	chromeManager  types.ChromeManager
	streamManager  *ws.StreamManager
	logger         *zap.Logger
	router         *gin.Engine
	httpServer     *http.Server
	config         *Config
	upgrader       websocket.Upgrader
}

// Config holds server configuration
type Config struct {
	Port           int    `json:"port"`
	Host           string `json:"host"`
	DefaultFormat  string `json:"default_format"`
	Quality        int    `json:"quality"`
	IncludeCursor  bool   `json:"include_cursor"`
	LogLevel       string `json:"log_level"`
	ChromeTimeout  string `json:"chrome_timeout"`
	// WebSocket streaming configuration
	StreamMaxSessions int `json:"stream_max_sessions"`
	StreamDefaultFPS  int `json:"stream_default_fps"`
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:              8080,
		Host:              "localhost",
		DefaultFormat:     "png",
		Quality:           95,
		IncludeCursor:     false,
		LogLevel:          "info",
		ChromeTimeout:     "30s",
		StreamMaxSessions: 10,
		StreamDefaultFPS:  10,
	}
}

// NewServer creates a new screenshot server
func NewServer() (*Server, error) {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize screenshot engine
	engine, err := screenshot.NewEngine()
	if err != nil {
		logger.Error("Failed to create screenshot engine", zap.Error(err))
		return nil, fmt.Errorf("failed to create screenshot engine: %w", err)
	}

	// Initialize Chrome manager
	chromeManager := chrome.NewManager()

	// Initialize stream manager
	streamManager := ws.NewStreamManager(logger)

	// Create WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// Create server instance
	server := &Server{
		engine:        engine,
		chromeManager: chromeManager,
		streamManager: streamManager,
		logger:        logger,
		config:        DefaultConfig(),
		upgrader:      upgrader,
	}

	// Setup HTTP router
	server.setupRouter()

	return server, nil
}

// setupRouter configures the HTTP routes
func (s *Server) setupRouter() {
	// Use gin in release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	s.router = gin.New()
	
	// Middleware
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
	s.router.Use(s.corsMiddleware())

	// Health check
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/v1")
	{
		// Screenshot endpoints
		v1.POST("/screenshot", s.takeScreenshot)
		v1.GET("/screenshot", s.takeScreenshotGET)
		
		// Window management
		v1.GET("/windows", s.listWindows)
		v1.GET("/windows/:handle", s.getWindow)
		
		// Chrome integration
		v1.GET("/chrome/instances", s.listChromeInstances)
		v1.GET("/chrome/tabs", s.listChromeTabs)
		v1.POST("/chrome/tabs/:id/screenshot", s.takeChromeTabScreenshot)
		
		// WebSocket streaming
		v1.GET("/stream/:windowId", s.handleWebSocketStream)
		v1.GET("/stream/status", s.getStreamStatus)
	}

	// API routes (for compatibility)
	api := s.router.Group("/api")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/windows", s.listWindows)
		api.GET("/screenshot", s.takeScreenshotGET)
	}

	// WebSocket streaming routes (top level for simplicity)
	s.router.GET("/stream/:windowId", s.handleWebSocketStream)

	// MCP JSON-RPC 2.0 endpoint
	s.router.POST("/rpc", s.handleMCPRequest)

	// Documentation
	s.router.Static("/docs", "./docs")
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler: s.router,
	}

	s.logger.Info("Starting screenshot MCP server",
		zap.String("address", s.httpServer.Addr),
		zap.String("version", "1.0.0"),
	)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	s.logger.Info("Server exited")
	return nil
}

// HTTP Handlers

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// takeScreenshot handles screenshot requests
func (s *Server) takeScreenshot(c *gin.Context) {
	var req types.ScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	s.processScreenshotRequest(c, &req)
}

// takeScreenshotGET handles GET screenshot requests
func (s *Server) takeScreenshotGET(c *gin.Context) {
	req := types.ScreenshotRequest{
		Method:  c.DefaultQuery("method", "title"),
		Target:  c.Query("target"),
		Format:  types.ImageFormat(c.DefaultQuery("format", s.config.DefaultFormat)),
		Quality: s.config.Quality,
	}

	if qualityStr := c.Query("quality"); qualityStr != "" {
		if quality, err := strconv.Atoi(qualityStr); err == nil {
			req.Quality = quality
		}
	}

	req.IncludeCursor = c.Query("cursor") == "true"

	if req.Target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target parameter is required"})
		return
	}

	s.processScreenshotRequest(c, &req)
}

// processScreenshotRequest processes a screenshot request
func (s *Server) processScreenshotRequest(c *gin.Context, req *types.ScreenshotRequest) {
	startTime := time.Now()

	options := &types.CaptureOptions{
		IncludeCursor:    req.IncludeCursor,
		IncludeFrame:     true,
		ScaleFactor:      1.0,
		AllowMinimized:   true,
		RestoreWindow:    false,
		WaitForVisible:   2 * time.Second,
		RetryCount:       3,
		CustomProperties: make(map[string]string),
	}

	if req.Region != nil {
		options.Region = req.Region
	}

	var buffer *types.ScreenshotBuffer
	var err error

	// Capture based on method
	switch req.Method {
	case "title":
		buffer, err = s.engine.CaptureByTitle(req.Target, options)
	case "pid":
		if pid, parseErr := strconv.ParseUint(req.Target, 10, 32); parseErr == nil {
			buffer, err = s.engine.CaptureByPID(uint32(pid), options)
		} else {
			err = fmt.Errorf("invalid PID: %s", req.Target)
		}
	case "handle":
		if handle, parseErr := strconv.ParseUint(req.Target, 10, 64); parseErr == nil {
			buffer, err = s.engine.CaptureByHandle(uintptr(handle), options)
		} else {
			err = fmt.Errorf("invalid handle: %s", req.Target)
		}
	case "class":
		buffer, err = s.engine.CaptureByClassName(req.Target, options)
	default:
		err = fmt.Errorf("unsupported method: %s", req.Method)
	}

	if err != nil {
		s.logger.Error("Screenshot capture failed",
			zap.String("method", req.Method),
			zap.String("target", req.Target),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Encode the image data as base64
	imageData := base64.StdEncoding.EncodeToString(buffer.Data)

	response := types.ScreenshotResponse{
		Success:   true,
		Data:      imageData,
		Format:    buffer.Format,
		Width:     buffer.Width,
		Height:    buffer.Height,
		Size:      int64(len(buffer.Data)),
		Timestamp: buffer.Timestamp,
		Metadata: types.Metadata{
			CaptureMethod:  req.Method,
			ProcessingTime: time.Since(startTime),
			WindowVisible:  buffer.WindowInfo.IsVisible,
			WindowMinimized: buffer.WindowInfo.State == "minimized",
			DPIScaling:     float64(buffer.DPI) / 96.0,
			ColorDepth:     32,
			Properties:     options.CustomProperties,
		},
	}

	s.logger.Info("Screenshot captured successfully",
		zap.String("method", req.Method),
		zap.String("target", req.Target),
		zap.Int("width", buffer.Width),
		zap.Int("height", buffer.Height),
		zap.Duration("processing_time", response.Metadata.ProcessingTime),
	)

	c.JSON(http.StatusOK, response)
}

// listWindows lists all available windows
func (s *Server) listWindows(c *gin.Context) {
	// For now return a placeholder - window enumeration can be implemented later
	c.JSON(http.StatusOK, gin.H{
		"windows": []interface{}{},
		"message": "Window enumeration will be implemented in a future version",
	})
}

// getWindow gets information about a specific window
func (s *Server) getWindow(c *gin.Context) {
	handle := c.Param("handle")
	c.JSON(http.StatusOK, gin.H{
		"handle":  handle,
		"message": "Window details not yet implemented",
	})
}

// listChromeInstances lists all Chrome instances
func (s *Server) listChromeInstances(c *gin.Context) {
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		s.logger.Error("Failed to discover Chrome instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// listChromeTabs lists tabs for all or specific Chrome instances
func (s *Server) listChromeTabs(c *gin.Context) {
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		s.logger.Error("Failed to discover Chrome instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var allTabs []types.ChromeTab
	for _, instance := range instances {
		tabs, err := s.chromeManager.GetTabs(&instance)
		if err != nil {
			s.logger.Warn("Failed to get tabs for Chrome instance",
				zap.Uint32("pid", instance.PID),
				zap.Error(err),
			)
			continue
		}
		allTabs = append(allTabs, tabs...)
	}

	c.JSON(http.StatusOK, gin.H{
		"tabs":  allTabs,
		"count": len(allTabs),
	})
}

// takeChromeTabScreenshot takes a screenshot of a specific Chrome tab
func (s *Server) takeChromeTabScreenshot(c *gin.Context) {
	tabID := c.Param("id")

	// Find the tab
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var targetTab *types.ChromeTab
	for _, instance := range instances {
		tabs, err := s.chromeManager.GetTabs(&instance)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Tab not found"})
		return
	}

	// Capture screenshot
	options := types.DefaultCaptureOptions()
	buffer, err := s.chromeManager.CaptureTab(targetTab, options)
	if err != nil {
		s.logger.Error("Failed to capture Chrome tab screenshot",
			zap.String("tab_id", tabID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Encode as base64
	imageData := base64.StdEncoding.EncodeToString(buffer.Data)

	response := types.ScreenshotResponse{
		Success:   true,
		Data:      imageData,
		Format:    buffer.Format,
		Width:     buffer.Width,
		Height:    buffer.Height,
		Size:      int64(len(buffer.Data)),
		Timestamp: buffer.Timestamp,
		Metadata: types.Metadata{
			CaptureMethod: "chrome_tab",
			Properties: map[string]string{
				"tab_id":    tabID,
				"tab_title": targetTab.Title,
				"tab_url":   targetTab.URL,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleMCPRequest handles MCP JSON-RPC 2.0 requests
func (s *Server) handleMCPRequest(c *gin.Context) {
	var req types.MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendMCPError(c, nil, -32700, "Parse error", nil)
		return
	}

	s.logger.Debug("Received MCP request",
		zap.String("method", req.Method),
		zap.Any("id", req.ID),
	)

	switch req.Method {
	case "screenshot.capture":
		s.handleMCPScreenshot(c, &req)
	case "window.list":
		s.handleMCPWindowList(c, &req)
	case "chrome.instances":
		s.handleMCPChromeInstances(c, &req)
	case "chrome.tabs":
		s.handleMCPChromeTabs(c, &req)
	case "chrome.tabCapture":
		s.handleMCPChromeTabCapture(c, &req)
	case "stream.status":
		s.handleMCPStreamStatus(c, &req)
	default:
		s.sendMCPError(c, req.ID, -32601, "Method not found", nil)
	}
}

// handleMCPScreenshot handles MCP screenshot requests
func (s *Server) handleMCPScreenshot(c *gin.Context, req *types.MCPRequest) {
	// Parse parameters
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		s.sendMCPError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	// Build screenshot request
	screenshotReq := types.ScreenshotRequest{
		Method:        getString(params, "method", "title"),
		Target:        getString(params, "target", ""),
		Format:        types.ImageFormat(getString(params, "format", s.config.DefaultFormat)),
		Quality:       getInt(params, "quality", s.config.Quality),
		IncludeCursor: getBool(params, "include_cursor", s.config.IncludeCursor),
	}

	if screenshotReq.Target == "" {
		s.sendMCPError(c, req.ID, -32602, "Missing required parameter: target", nil)
		return
	}

	// Process the request (reuse existing logic)
	options := &types.CaptureOptions{
		IncludeCursor:    screenshotReq.IncludeCursor,
		IncludeFrame:     getBool(params, "include_frame", true),
		ScaleFactor:      getFloat64(params, "scale_factor", 1.0),
		AllowMinimized:   getBool(params, "allow_minimized", true),
		RestoreWindow:    getBool(params, "restore_window", false),
		WaitForVisible:   2 * time.Second,
		RetryCount:       3,
		CustomProperties: make(map[string]string),
	}

	var buffer *types.ScreenshotBuffer
	var err error

	switch screenshotReq.Method {
	case "title":
		buffer, err = s.engine.CaptureByTitle(screenshotReq.Target, options)
	case "pid":
		if pid, parseErr := strconv.ParseUint(screenshotReq.Target, 10, 32); parseErr == nil {
			buffer, err = s.engine.CaptureByPID(uint32(pid), options)
		} else {
			err = fmt.Errorf("invalid PID: %s", screenshotReq.Target)
		}
	case "handle":
		if handle, parseErr := strconv.ParseUint(screenshotReq.Target, 10, 64); parseErr == nil {
			buffer, err = s.engine.CaptureByHandle(uintptr(handle), options)
		} else {
			err = fmt.Errorf("invalid handle: %s", screenshotReq.Target)
		}
	case "class":
		buffer, err = s.engine.CaptureByClassName(screenshotReq.Target, options)
	default:
		err = fmt.Errorf("unsupported method: %s", screenshotReq.Method)
	}

	if err != nil {
		s.sendMCPError(c, req.ID, -32603, "Internal error", err.Error())
		return
	}

	// Encode and send response
	imageData := base64.StdEncoding.EncodeToString(buffer.Data)
	result := types.ScreenshotResponse{
		Success:   true,
		Data:      imageData,
		Format:    buffer.Format,
		Width:     buffer.Width,
		Height:    buffer.Height,
		Size:      int64(len(buffer.Data)),
		Timestamp: buffer.Timestamp,
	}

	s.sendMCPResult(c, req.ID, result)
}

// handleMCPWindowList handles MCP window list requests
func (s *Server) handleMCPWindowList(c *gin.Context, req *types.MCPRequest) {
	// Placeholder implementation
	result := map[string]interface{}{
		"windows": []interface{}{},
		"message": "Window enumeration not yet implemented",
	}
	s.sendMCPResult(c, req.ID, result)
}

// handleMCPChromeInstances handles MCP Chrome instances requests
func (s *Server) handleMCPChromeInstances(c *gin.Context, req *types.MCPRequest) {
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		s.sendMCPError(c, req.ID, -32603, "Internal error", err.Error())
		return
	}

	result := map[string]interface{}{
		"instances": instances,
		"count":     len(instances),
	}
	s.sendMCPResult(c, req.ID, result)
}

// handleMCPChromeTabs handles MCP Chrome tabs requests
func (s *Server) handleMCPChromeTabs(c *gin.Context, req *types.MCPRequest) {
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		s.sendMCPError(c, req.ID, -32603, "Internal error", err.Error())
		return
	}

	var allTabs []types.ChromeTab
	for _, instance := range instances {
		tabs, err := s.chromeManager.GetTabs(&instance)
		if err != nil {
			continue
		}
		allTabs = append(allTabs, tabs...)
	}

	result := map[string]interface{}{
		"tabs":  allTabs,
		"count": len(allTabs),
	}
	s.sendMCPResult(c, req.ID, result)
}

// handleMCPChromeTabCapture handles MCP Chrome tab capture requests
func (s *Server) handleMCPChromeTabCapture(c *gin.Context, req *types.MCPRequest) {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		s.sendMCPError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	tabID := getString(params, "tab_id", "")
	if tabID == "" {
		s.sendMCPError(c, req.ID, -32602, "Missing required parameter: tab_id", nil)
		return
	}

	// Find the tab (reuse existing logic)
	instances, err := s.chromeManager.DiscoverInstances()
	if err != nil {
		s.sendMCPError(c, req.ID, -32603, "Internal error", err.Error())
		return
	}

	var targetTab *types.ChromeTab
	for _, instance := range instances {
		tabs, err := s.chromeManager.GetTabs(&instance)
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
		s.sendMCPError(c, req.ID, -32603, "Tab not found", nil)
		return
	}

	// Capture screenshot
	options := types.DefaultCaptureOptions()
	buffer, err := s.chromeManager.CaptureTab(targetTab, options)
	if err != nil {
		s.sendMCPError(c, req.ID, -32603, "Screenshot failed", err.Error())
		return
	}

	// Encode and send response
	imageData := base64.StdEncoding.EncodeToString(buffer.Data)
	result := types.ScreenshotResponse{
		Success:   true,
		Data:      imageData,
		Format:    buffer.Format,
		Width:     buffer.Width,
		Height:    buffer.Height,
		Size:      int64(len(buffer.Data)),
		Timestamp: buffer.Timestamp,
	}

	s.sendMCPResult(c, req.ID, result)
}

// MCP helper functions

func (s *Server) sendMCPResult(c *gin.Context, id interface{}, result interface{}) {
	response := types.MCPResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) sendMCPError(c *gin.Context, id interface{}, code int, message string, data interface{}) {
	response := types.MCPResponse{
		JSONRPC: "2.0",
		Error: &types.MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
	c.JSON(http.StatusOK, response) // MCP errors are still HTTP 200
}

// Parameter parsing helpers
func getString(params map[string]interface{}, key string, defaultValue string) string {
	if val, exists := params[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getInt(params map[string]interface{}, key string, defaultValue int) int {
	if val, exists := params[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func getBool(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := params[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getFloat64(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, exists := params[key]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

// Middleware

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		s.logger.Info("HTTP Request",
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// handleMCPStreamStatus handles MCP stream status requests
func (s *Server) handleMCPStreamStatus(c *gin.Context, req *types.MCPRequest) {
	stats := s.streamManager.GetStats()
	result := map[string]interface{}{
		"active_sessions": stats.ActiveSessions,
		"total_sessions":  stats.TotalSessions,
		"total_frames":    stats.TotalFrames,
		"uptime":          stats.Uptime.String(),
		"max_sessions":    s.config.StreamMaxSessions,
		"websocket_url":   fmt.Sprintf("ws://%s:%d/stream/{windowId}", s.config.Host, s.config.Port),
	}
	s.sendMCPResult(c, req.ID, result)
}

// WebSocket streaming handlers

// handleWebSocketStream handles WebSocket streaming connections
func (s *Server) handleWebSocketStream(c *gin.Context) {
	windowIDStr := c.Param("windowId")
	windowID, err := strconv.Atoi(windowIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid window ID"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// Parse query parameters for initial options
	fps := s.config.StreamDefaultFPS
	quality := s.config.Quality
	format := s.config.DefaultFormat

	if fpsStr := c.Query("fps"); fpsStr != "" {
		if f, err := strconv.Atoi(fpsStr); err == nil && f > 0 && f <= 60 {
			fps = f
		}
	}

	if qualityStr := c.Query("quality"); qualityStr != "" {
		if q, err := strconv.Atoi(qualityStr); err == nil && q > 0 && q <= 100 {
			quality = q
		}
	}

	if formatStr := c.Query("format"); formatStr != "" {
		format = formatStr
	}

	options := &types.StreamOptions{
		FPS:      fps,
		Quality:  quality,
		Format:   types.ImageFormat(format),
	}

	// Set up the screenshot engine in the stream manager
	s.streamManager.SetEngine(s.engine)

	s.logger.Info("Starting WebSocket stream session",
		zap.Int("window_id", windowID),
		zap.Int("fps", fps),
		zap.Int("quality", quality),
		zap.String("format", format),
		zap.String("client_ip", c.ClientIP()),
	)

	// Special handling: if windowID is 0, capture full desktop
	if windowID == 0 {
		s.logger.Info("Using desktop capture mode for window ID 0")
	}

	// Start streaming session
	session, err := s.streamManager.StartSession(uintptr(windowID), options)
	if err != nil {
		s.logger.Error("Stream session failed",
			zap.Int("window_id", windowID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the WebSocket connection
	session.Conn = conn

	// Send session started message
	err = conn.WriteJSON(map[string]interface{}{
		"type":       "session_started",
		"session_id": session.ID,
		"timestamp":  time.Now(),
	})
	if err != nil {
		s.logger.Error("Failed to send session started message", zap.Error(err))
		return
	}

	// Handle WebSocket messages in a goroutine
	go s.streamManager.HandleClientMessages(session)

	// Wait for session to complete
	<-session.Context.Done()

	s.logger.Info("WebSocket stream session ended",
		zap.Int("window_id", windowID),
		zap.String("client_ip", c.ClientIP()),
	)
}

// getStreamStatus returns the current streaming status
func (s *Server) getStreamStatus(c *gin.Context) {
	stats := s.streamManager.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"active_sessions": stats.ActiveSessions,
		"total_sessions":  stats.TotalSessions,
		"total_frames":    stats.TotalFrames,
		"uptime":          stats.Uptime.String(),
		"max_sessions":    s.config.StreamMaxSessions,
	})
}

// main function
func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
