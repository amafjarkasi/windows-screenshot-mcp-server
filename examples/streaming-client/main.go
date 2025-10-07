package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Message types and structures matching the server implementation
type MessageType string

const (
	TypeSessionStarted MessageType = "session_started"
	TypeFrame          MessageType = "frame"
	TypeSessionUpdated MessageType = "session_updated"
	TypeStatus         MessageType = "status"
	TypeError          MessageType = "error"
)

type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type FrameData struct {
	DataURL   string `json:"data_url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size      int    `json:"size"`
	Timestamp int64  `json:"timestamp"`
	Format    string `json:"format"`
}

type ControlMessage struct {
	Command string                 `json:"command"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type StreamingClient struct {
	serverURL string
	windowID  int
	conn      *websocket.Conn
	done      chan struct{}
	interrupt chan os.Signal
}

func NewStreamingClient(serverURL string, windowID int) *StreamingClient {
	return &StreamingClient{
		serverURL: serverURL,
		windowID:  windowID,
		done:      make(chan struct{}),
		interrupt: make(chan os.Signal, 1),
	}
}

func (c *StreamingClient) Connect() error {
	u := url.URL{
		Scheme: "ws",
		Host:   c.serverURL,
		Path:   fmt.Sprintf("/stream/%d", c.windowID),
	}

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *StreamingClient) Listen() {
	defer close(c.done)

	for {
		select {
		case <-c.interrupt:
			log.Println("Interrupt received, stopping...")
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}

			var wsMsg WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("JSON unmarshal error: %v", err)
				continue
			}

			c.handleMessage(wsMsg)
		}
	}
}

func (c *StreamingClient) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case TypeSessionStarted:
		log.Printf("✅ Session started: %s", msg.SessionID)
		
	case TypeFrame:
		frameBytes, _ := json.Marshal(msg.Data)
		var frame FrameData
		if err := json.Unmarshal(frameBytes, &frame); err != nil {
			log.Printf("Error parsing frame data: %v", err)
			return
		}
		
		log.Printf("📸 Frame received: %dx%d, %s, %d bytes", 
			frame.Width, frame.Height, frame.Format, frame.Size)
			
		// You can save the frame or process it here
		// The frame.DataURL contains the base64-encoded image data
		
	case TypeSessionUpdated:
		log.Printf("🔄 Session settings updated")
		
	case TypeStatus:
		statusBytes, _ := json.Marshal(msg.Data)
		log.Printf("ℹ️  Status update: %s", string(statusBytes))
		
	case TypeError:
		log.Printf("❌ Error: %s", msg.Error)
		
	default:
		log.Printf("⚠️  Unknown message type: %s", msg.Type)
	}
}

func (c *StreamingClient) UpdateOptions(fps int, quality int, format string) error {
	controlMsg := ControlMessage{
		Command: "update_options",
		Options: map[string]interface{}{
			"fps":     fps,
			"quality": quality,
			"format":  format,
		},
	}

	return c.conn.WriteJSON(controlMsg)
}

func (c *StreamingClient) RequestStatus() error {
	controlMsg := ControlMessage{
		Command: "status",
	}

	return c.conn.WriteJSON(controlMsg)
}

func (c *StreamingClient) Stop() error {
	controlMsg := ControlMessage{
		Command: "stop",
	}

	if err := c.conn.WriteJSON(controlMsg); err != nil {
		return err
	}

	// Close connection gracefully
	return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (c *StreamingClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *StreamingClient) Run(ctx context.Context) error {
	signal.Notify(c.interrupt, os.Interrupt, syscall.SIGTERM)

	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Close()

	go c.Listen()

	// Example: Update options after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("🔧 Updating stream options...")
		if err := c.UpdateOptions(15, 90, "png"); err != nil {
			log.Printf("Failed to update options: %v", err)
		}
	}()

	// Example: Request status after 10 seconds
	go func() {
		time.Sleep(10 * time.Second)
		log.Println("📊 Requesting status...")
		if err := c.RequestStatus(); err != nil {
			log.Printf("Failed to request status: %v", err)
		}
	}()

	// Wait for interrupt or completion
	select {
	case <-c.interrupt:
		log.Println("Shutting down gracefully...")
		if err := c.Stop(); err != nil {
			log.Printf("Error stopping stream: %v", err)
		}
	case <-c.done:
		log.Println("Stream ended")
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	return nil
}

func parseWindowList(windowListStr string) ([]int, error) {
	if windowListStr == "" {
		return nil, fmt.Errorf("no window IDs provided")
	}

	parts := strings.Split(windowListStr, ",")
	windowIDs := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid window ID '%s': %w", part, err)
		}

		windowIDs = append(windowIDs, id)
	}

	if len(windowIDs) == 0 {
		return nil, fmt.Errorf("no valid window IDs found")
	}

	return windowIDs, nil
}

func main() {
	var (
		serverURL = flag.String("server", "localhost:8080", "Server URL (host:port)")
		windowIDs = flag.String("windows", "", "Comma-separated list of window IDs to stream")
		timeout   = flag.Duration("timeout", 30*time.Second, "Connection timeout")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	if !*verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if *windowIDs == "" {
		log.Fatal("Window ID(s) must be specified with -windows flag")
	}

	windows, err := parseWindowList(*windowIDs)
	if err != nil {
		log.Fatalf("Error parsing window IDs: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	log.Printf("🚀 Starting streaming client for windows: %v", windows)
	log.Printf("📡 Server: %s", *serverURL)
	log.Printf("⏱️  Timeout: %s", *timeout)

	// For simplicity, we'll just stream the first window
	// In a real application, you might want to handle multiple streams
	windowID := windows[0]
	if len(windows) > 1 {
		log.Printf("⚠️  Multiple windows specified, using first one: %d", windowID)
	}

	client := NewStreamingClient(*serverURL, windowID)
	
	if err := client.Run(ctx); err != nil {
		log.Fatalf("Client error: %v", err)
	}

	log.Println("✅ Client finished")
}