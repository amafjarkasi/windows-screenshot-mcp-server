package ws

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/screenshot-mcp-server/internal/screenshot"
	"github.com/screenshot-mcp-server/pkg/types"
	"go.uber.org/zap"
)

// StreamManager manages WebSocket streaming sessions
type StreamManager struct {
	sessions    map[string]*StreamSession
	sessionsMux sync.RWMutex
	upgrader    websocket.Upgrader
	engine      types.ScreenshotEngine
	processor   types.ImageProcessor
	logger      *zap.Logger
}

// StreamSession represents an active streaming session
type StreamSession struct {
	ID          string                    `json:"id"`
	WindowID    uintptr                   `json:"window_id"`
	Conn        *websocket.Conn           `json:"-"`
	Options     *types.StreamOptions      `json:"options"`
	Active      bool                      `json:"active"`
	StartTime   time.Time                 `json:"start_time"`
	FrameCount  int64                     `json:"frame_count"`
	BytesSent   int64                     `json:"bytes_sent"`
	LastFrame   time.Time                 `json:"last_frame"`
	StopChan    chan struct{}             `json:"-"`
	Context     context.Context           `json:"-"`
	Cancel      context.CancelFunc        `json:"-"`
	ClientInfo  *ClientInfo               `json:"client_info"`
	mutex       sync.RWMutex
}

// ClientInfo contains information about the connected client
type ClientInfo struct {
	RemoteAddr string            `json:"remote_addr"`
	UserAgent  string            `json:"user_agent"`
	Headers    map[string]string `json:"headers"`
	ConnectedAt time.Time        `json:"connected_at"`
}

// StreamMessage represents a message sent over WebSocket
type StreamMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// FrameMessage contains screenshot frame data
type FrameMessage struct {
	FrameNumber int64  `json:"frame_number"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Format      string `json:"format"`
	DataURL     string `json:"data_url"` // Base64 encoded image as data URL
	Size        int    `json:"size"`
	Timestamp   time.Time `json:"timestamp"`
}

// StatusMessage contains session status information
type StatusMessage struct {
	SessionID   string               `json:"session_id"`
	WindowID    uintptr              `json:"window_id"`
	Active      bool                 `json:"active"`
	FPS         int                  `json:"fps"`
	FrameCount  int64                `json:"frame_count"`
	BytesSent   int64                `json:"bytes_sent"`
	Duration    time.Duration        `json:"duration"`
	Options     *types.StreamOptions `json:"options"`
}

// ControlMessage represents control commands
type ControlMessage struct {
	Command   string                   `json:"command"`
	SessionID string                   `json:"session_id,omitempty"`
	Options   *types.StreamOptions     `json:"options,omitempty"`
	WindowID  *uintptr                 `json:"window_id,omitempty"`
}

// NewStreamManager creates a new stream manager
func NewStreamManager(logger *zap.Logger) *StreamManager {
	processor := screenshot.NewImageProcessor()
	
	return &StreamManager{
		sessions: make(map[string]*StreamSession),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024 * 1024, // 1MB buffer for large frames
		},
		processor: processor,
		logger:    logger,
	}
}

// HandleWebSocket handles WebSocket connections for streaming
func (sm *StreamManager) HandleWebSocket(c *gin.Context) {
	// Extract parameters
	windowIDStr := c.Param("windowId")
	if windowIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "window_id parameter is required"})
		return
	}

	// Parse window ID
	var windowID uintptr
	if _, err := fmt.Sscanf(windowIDStr, "%d", &windowID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window_id"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := sm.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		sm.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	// Create client info
	clientInfo := &ClientInfo{
		RemoteAddr:  conn.RemoteAddr().String(),
		UserAgent:   c.Request.Header.Get("User-Agent"),
		ConnectedAt: time.Now(),
		Headers:     make(map[string]string),
	}
	
	// Copy relevant headers
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			clientInfo.Headers[key] = values[0]
		}
	}

	// Start streaming session
	sessionID := fmt.Sprintf("stream_%d_%d", windowID, time.Now().Unix())
	options := types.DefaultStreamOptions()
	
	session, err := sm.StartSession(windowID, options)
	if err != nil {
		conn.WriteJSON(StreamMessage{
			Type:      "error",
			Timestamp: time.Now(),
			Error:     err.Error(),
		})
		conn.Close()
		return
	}

	// Update session with WebSocket connection and client info
	sm.sessionsMux.Lock()
	if existingSession, exists := sm.sessions[sessionID]; exists {
		existingSession.Conn = conn
		existingSession.ClientInfo = clientInfo
		session = existingSession
	}
	sm.sessionsMux.Unlock()

	sm.logger.Info("WebSocket streaming session started",
		zap.String("session_id", sessionID),
		zap.Uintptr("window_id", windowID),
		zap.String("client_addr", clientInfo.RemoteAddr),
	)

	// Send initial status message
	conn.WriteJSON(StreamMessage{
		Type:      "session_started",
		Timestamp: time.Now(),
		SessionID: sessionID,
		Data: StatusMessage{
			SessionID: sessionID,
			WindowID:  windowID,
			Active:    true,
			FPS:       options.FPS,
			Options:   options,
		},
	})

	// Handle incoming messages
	go sm.handleClientMessages(session)

	// Wait for session to end
	<-session.Context.Done()
	
	sm.logger.Info("WebSocket streaming session ended",
		zap.String("session_id", sessionID),
		zap.Int64("frames_sent", session.FrameCount),
		zap.Int64("bytes_sent", session.BytesSent),
	)
	
	conn.Close()
}

// StartSession starts a new streaming session
func (sm *StreamManager) StartSession(windowID uintptr, options *types.StreamOptions) (*StreamSession, error) {
	if options == nil {
		options = types.DefaultStreamOptions()
	}

	sessionID := fmt.Sprintf("stream_%d_%d", windowID, time.Now().UnixNano())
	
	ctx, cancel := context.WithCancel(context.Background())
	
	session := &StreamSession{
		ID:        sessionID,
		WindowID:  windowID,
		Options:   options,
		Active:    true,
		StartTime: time.Now(),
		StopChan:  make(chan struct{}),
		Context:   ctx,
		Cancel:    cancel,
	}

	// Store session
	sm.sessionsMux.Lock()
	sm.sessions[sessionID] = session
	sm.sessionsMux.Unlock()

	// Start streaming goroutine
	go sm.streamFrames(session)

	sm.logger.Info("Streaming session started",
		zap.String("session_id", sessionID),
		zap.Uintptr("window_id", windowID),
		zap.Int("fps", options.FPS),
	)

	return session, nil
}

// StopSession stops a streaming session
func (sm *StreamManager) StopSession(sessionID string) error {
	sm.sessionsMux.Lock()
	defer sm.sessionsMux.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Mark as inactive and cancel context
	session.mutex.Lock()
	session.Active = false
	session.Cancel()
	session.mutex.Unlock()

	// Remove from active sessions
	delete(sm.sessions, sessionID)

	sm.logger.Info("Streaming session stopped",
		zap.String("session_id", sessionID),
		zap.Int64("frames_sent", session.FrameCount),
	)

	return nil
}

// GetActiveSessions returns all active streaming sessions
func (sm *StreamManager) GetActiveSessions() ([]*types.StreamSession, error) {
	sm.sessionsMux.RLock()
	defer sm.sessionsMux.RUnlock()

	var sessions []*types.StreamSession
	for _, session := range sm.sessions {
		if session.Active {
			sessions = append(sessions, &types.StreamSession{
				ID:         session.ID,
				WindowID:   session.WindowID,
				FPS:        session.Options.FPS,
				Quality:    session.Options.Quality,
				Format:     session.Options.Format,
				Active:     session.Active,
				StartTime:  session.StartTime,
				FrameCount: session.FrameCount,
				BytesSent:  session.BytesSent,
			})
		}
	}

	return sessions, nil
}

// UpdateSession updates session parameters
func (sm *StreamManager) UpdateSession(sessionID string, options *types.StreamOptions) error {
	sm.sessionsMux.RLock()
	session, exists := sm.sessions[sessionID]
	sm.sessionsMux.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Update options
	session.mutex.Lock()
	if options.FPS > 0 {
		session.Options.FPS = options.FPS
	}
	if options.Quality > 0 {
		session.Options.Quality = options.Quality
	}
	if options.Format != "" {
		session.Options.Format = options.Format
	}
	if options.MaxWidth > 0 {
		session.Options.MaxWidth = options.MaxWidth
	}
	if options.MaxHeight > 0 {
		session.Options.MaxHeight = options.MaxHeight
	}
	session.mutex.Unlock()

	sm.logger.Info("Streaming session updated",
		zap.String("session_id", sessionID),
		zap.Int("fps", session.Options.FPS),
		zap.Int("quality", session.Options.Quality),
	)

	// Send status update to client
	if session.Conn != nil {
		session.Conn.WriteJSON(StreamMessage{
			Type:      "session_updated",
			Timestamp: time.Now(),
			SessionID: sessionID,
			Data: StatusMessage{
				SessionID: sessionID,
				WindowID:  session.WindowID,
				Active:    session.Active,
				FPS:       session.Options.FPS,
				Options:   session.Options,
			},
		})
	}

	return nil
}

// streamFrames continuously captures and streams frames
func (sm *StreamManager) streamFrames(session *StreamSession) {
	defer func() {
		if r := recover(); r != nil {
			sm.logger.Error("Streaming goroutine panicked",
				zap.String("session_id", session.ID),
				zap.Any("error", r),
			)
		}
	}()

	captureOptions := types.DefaultCaptureOptions()
	captureOptions.AllowMinimized = true
	captureOptions.RestoreWindow = false

	frameDuration := time.Duration(1000/session.Options.FPS) * time.Millisecond
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-session.Context.Done():
			return
		case <-ticker.C:
			if !session.Active {
				return
			}

			// Update ticker if FPS changed
			session.mutex.RLock()
			newFrameDuration := time.Duration(1000/session.Options.FPS) * time.Millisecond
			if newFrameDuration != frameDuration {
				frameDuration = newFrameDuration
				ticker.Reset(frameDuration)
			}
			currentOptions := *session.Options
			session.mutex.RUnlock()

			// Capture screenshot
			buffer, err := sm.engine.CaptureByHandle(session.WindowID, captureOptions)
			if err != nil {
				sm.logger.Warn("Failed to capture frame",
					zap.String("session_id", session.ID),
					zap.Error(err),
				)
				continue
			}

			// Process frame
			if err := sm.processAndSendFrame(session, buffer, &currentOptions); err != nil {
				sm.logger.Error("Failed to process frame",
					zap.String("session_id", session.ID),
					zap.Error(err),
				)
			}
		}
	}
}

// processAndSendFrame processes and sends a frame to the client
func (sm *StreamManager) processAndSendFrame(session *StreamSession, buffer *types.ScreenshotBuffer, options *types.StreamOptions) error {
	// Resize if needed
	if options.MaxWidth > 0 && buffer.Width > options.MaxWidth {
		aspectRatio := float64(buffer.Height) / float64(buffer.Width)
		newHeight := int(float64(options.MaxWidth) * aspectRatio)
		if newHeight > options.MaxHeight && options.MaxHeight > 0 {
			newHeight = options.MaxHeight
			options.MaxWidth = int(float64(newHeight) / aspectRatio)
		}
		
		resized, err := sm.processor.Resize(buffer, options.MaxWidth, newHeight)
		if err != nil {
			return fmt.Errorf("failed to resize frame: %w", err)
		}
		buffer = resized
	}

	// Encode frame
	encoded, err := sm.processor.Encode(buffer, options.Format, options.Quality)
	if err != nil {
		return fmt.Errorf("failed to encode frame: %w", err)
	}

	// Create data URL
	var mimeType string
	switch options.Format {
	case types.FormatPNG:
		mimeType = "image/png"
	case types.FormatJPEG:
		mimeType = "image/jpeg"
	case types.FormatWebP:
		mimeType = "image/webp"
	default:
		mimeType = "image/png"
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(encoded)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// Create frame message
	frame := FrameMessage{
		FrameNumber: session.FrameCount + 1,
		Width:       buffer.Width,
		Height:      buffer.Height,
		Format:      string(options.Format),
		DataURL:     dataURL,
		Size:        len(encoded),
		Timestamp:   time.Now(),
	}

	// Send frame to client
	if session.Conn != nil {
		err := session.Conn.WriteJSON(StreamMessage{
			Type:      "frame",
			Timestamp: time.Now(),
			SessionID: session.ID,
			Data:      frame,
		})
		
		if err != nil {
			return fmt.Errorf("failed to send frame: %w", err)
		}
	}

	// Update session stats
	session.mutex.Lock()
	session.FrameCount++
	session.BytesSent += int64(len(encoded))
	session.LastFrame = time.Now()
	session.mutex.Unlock()

	return nil
}

// handleClientMessages handles incoming WebSocket messages from clients
func (sm *StreamManager) handleClientMessages(session *StreamSession) {
	defer func() {
		if r := recover(); r != nil {
			sm.logger.Error("Client message handler panicked",
				zap.String("session_id", session.ID),
				zap.Any("error", r),
			)
		}
	}()

	for {
		select {
		case <-session.Context.Done():
			return
		default:
			var msg ControlMessage
			err := session.Conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					sm.logger.Error("WebSocket error",
						zap.String("session_id", session.ID),
						zap.Error(err),
					)
				}
				// Stop session on connection error
				sm.StopSession(session.ID)
				return
			}

			sm.handleControlMessage(session, &msg)
		}
	}
}

// handleControlMessage processes control messages from clients
func (sm *StreamManager) handleControlMessage(session *StreamSession, msg *ControlMessage) {
	switch msg.Command {
	case "update_options":
		if msg.Options != nil {
			err := sm.UpdateSession(session.ID, msg.Options)
			if err != nil {
				session.Conn.WriteJSON(StreamMessage{
					Type:      "error",
					Timestamp: time.Now(),
					SessionID: session.ID,
					Error:     err.Error(),
				})
			}
		}
		
	case "get_status":
		session.mutex.RLock()
		status := StatusMessage{
			SessionID:  session.ID,
			WindowID:   session.WindowID,
			Active:     session.Active,
			FPS:        session.Options.FPS,
			FrameCount: session.FrameCount,
			BytesSent:  session.BytesSent,
			Duration:   time.Since(session.StartTime),
			Options:    session.Options,
		}
		session.mutex.RUnlock()
		
		session.Conn.WriteJSON(StreamMessage{
			Type:      "status",
			Timestamp: time.Now(),
			SessionID: session.ID,
			Data:      status,
		})
		
	case "stop":
		sm.StopSession(session.ID)
		
	default:
		session.Conn.WriteJSON(StreamMessage{
			Type:      "error",
			Timestamp: time.Now(),
			SessionID: session.ID,
			Error:     fmt.Sprintf("unknown command: %s", msg.Command),
		})
	}
}

// GetSessionStats returns statistics for a session
func (sm *StreamManager) GetSessionStats(sessionID string) (*StatusMessage, error) {
	sm.sessionsMux.RLock()
	session, exists := sm.sessions[sessionID]
	sm.sessionsMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return &StatusMessage{
		SessionID:  session.ID,
		WindowID:   session.WindowID,
		Active:     session.Active,
		FPS:        session.Options.FPS,
		FrameCount: session.FrameCount,
		BytesSent:  session.BytesSent,
		Duration:   time.Since(session.StartTime),
		Options:    session.Options,
	}, nil
}

// Cleanup stops all sessions and cleans up resources
func (sm *StreamManager) Cleanup() {
	sm.sessionsMux.Lock()
	defer sm.sessionsMux.Unlock()

	for sessionID, session := range sm.sessions {
		session.Active = false
		session.Cancel()
		if session.Conn != nil {
			session.Conn.Close()
		}
		delete(sm.sessions, sessionID)
	}

	sm.logger.Info("Stream manager cleaned up")
}

// GetStats returns overall streaming statistics
func (sm *StreamManager) GetStats() *StreamStats {
	sm.sessionsMux.RLock()
	defer sm.sessionsMux.RUnlock()

	activeCount := 0
	totalFrames := int64(0)
	for _, session := range sm.sessions {
		if session.Active {
			activeCount++
		}
		session.mutex.RLock()
		totalFrames += session.FrameCount
		session.mutex.RUnlock()
	}

	return &StreamStats{
		ActiveSessions: activeCount,
		TotalSessions:  len(sm.sessions),
		TotalFrames:    totalFrames,
		Uptime:         time.Since(time.Now()), // This should be set when manager starts
	}
}

// StreamStats contains overall streaming statistics
type StreamStats struct {
	ActiveSessions int           `json:"active_sessions"`
	TotalSessions  int           `json:"total_sessions"`
	TotalFrames    int64         `json:"total_frames"`
	Uptime         time.Duration `json:"uptime"`
}

// SetEngine sets the screenshot engine (public method)
func (sm *StreamManager) SetEngine(engine types.ScreenshotEngine) {
	sm.engine = engine
}

// HandleClientMessages handles client messages (public method)
func (sm *StreamManager) HandleClientMessages(session *StreamSession) {
	sm.handleClientMessages(session)
}

// Note: StreamManager has a custom implementation that doesn't strictly follow
// the types.StreamManager interface to support WebSocket-specific functionality
