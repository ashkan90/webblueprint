package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"

	"github.com/gorilla/websocket"
)

// WebSocketManager handles WebSocket connections and messaging
type WebSocketManager struct {
	clients    map[string]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan []byte
	mutex      sync.RWMutex
	Logger     Logger // Logger interface for error handling
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	manager  *WebSocketManager
	conn     *websocket.Conn
	send     chan []byte
	clientID string
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Message types
const (
	MsgTypeNodeIntro    = "node.intro"       // Node type introduction
	MsgTypeNodeStart    = "node.start"       // Node execution started
	MsgTypeNodeComplete = "node.complete"    // Node execution completed
	MsgTypeNodeError    = "node.error"       // Node execution error
	MsgTypeDataFlow     = "data.flow"        // Data flowing between nodes
	MsgTypeDebugData    = "debug.data"       // Debug data available
	MsgTypeExecStart    = "execution.start"  // Blueprint execution started
	MsgTypeExecEnd      = "execution.end"    // Blueprint execution ended
	MsgTypeExecStatus   = "execution.status" // Execution status update
	MsgTypeResult       = "result"           // Pin output value
	MsgTypeLog          = "log"              // Log message
)

// HTTP connection upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	manager := &WebSocketManager{
		clients:    make(map[string]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan []byte),
	}

	go manager.run()
	return manager
}

// run handles client registration and broadcasting
func (h *WebSocketManager) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.clientID] = client
			h.mutex.Unlock()
			log.Printf("Client connected: %s", client.clientID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.clientID]; ok {
				delete(h.clients, client.clientID)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected: %s", client.clientID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Channel full, close connection
					close(client.send)
					delete(h.clients, client.clientID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// HandleWebSocket handles a new WebSocket connection
func (h *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create a new client
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &WebSocketClient{
		manager:  h,
		conn:     conn,
		send:     make(chan []byte, 256),
		clientID: clientID,
	}

	// Register client
	h.register <- client

	// Start client handlers
	go client.readPump()
	go client.writePump()

	// Send welcome message
	client.sendMessage(MsgTypeExecStatus, map[string]interface{}{
		"status":  "connected",
		"message": "WebSocket connection established",
	})
}

// BroadcastMessage sends a message to all connected clients
func (h *WebSocketManager) BroadcastMessage(messageType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling message payload: %v", err)
		return
	}

	msg := WebSocketMessage{
		Type:    messageType,
		Payload: data,
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- msgData
}

// SendErrorNotification sends an error notification to clients
func (h *WebSocketManager) SendErrorNotification(executionID string, err *bperrors.BlueprintError) {
	notification := ErrorNotification{
		Type:        "error",
		Error:       err,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeNodeError, notification)
}

// SendErrorAnalysisNotification sends an error analysis notification to clients
func (h *WebSocketManager) SendErrorAnalysisNotification(executionID string, analysis map[string]interface{}) {
	notification := ErrorAnalysisNotification{
		Type:        "error_analysis",
		Analysis:    analysis,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeDebugData, notification)
}

// SendRecoveryNotification sends a recovery notification to clients
func (h *WebSocketManager) SendRecoveryNotification(executionID, nodeID, errorCode, strategy string, successful bool, details map[string]interface{}) {
	notification := RecoveryNotification{
		Type:        "recovery_attempt",
		Successful:  successful,
		Strategy:    strategy,
		NodeID:      nodeID,
		ErrorCode:   errorCode,
		Details:     details,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeExecStatus, notification)
}

// readPump handles messages from the client
func (c *WebSocketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(1024 * 1024) // 1MB max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the message
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle message based on type
		// This will be expanded as we add more message handlers
		switch wsMsg.Type {
		case "blueprint.execute":
			// Handle blueprint execution request
			log.Printf("Blueprint execution requested")

		case "node.inspect":
			// Handle node inspection request
			log.Printf("Node inspection requested")
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel was closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendMessage sends a message to the client
func (c *WebSocketClient) sendMessage(messageType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling message payload: %v", err)
		return
	}

	msg := WebSocketMessage{
		Type:    messageType,
		Payload: data,
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	c.send <- msgData
}

// Create a WebSocket logger that sends logs to clients
type WebSocketLogger struct {
	wsManager        *WebSocketManager
	persistentLogger *os.File
	nodeID           string
	mu               sync.RWMutex
}

// NewWebSocketLogger creates a new logger that sends logs via WebSocket
func NewWebSocketLogger(wsManager *WebSocketManager) *WebSocketLogger {
	f, err := os.OpenFile("./log.out", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v, skipping persistent logger", err)
	}

	return &WebSocketLogger{
		wsManager:        wsManager,
		persistentLogger: f,
	}
}

func (l *WebSocketLogger) Opts(m map[string]interface{}) {
	for s, i := range m {
		if s == "nodeId" {
			l.nodeID = i.(string)
		}
	}
}

// Implement Logger interface methods
func (l *WebSocketLogger) Debug(msg string, fields map[string]interface{}) {
	l.sendLogMessage("debug", msg, fields)
}

func (l *WebSocketLogger) Info(msg string, fields map[string]interface{}) {
	l.sendLogMessage("info", msg, fields)
}

func (l *WebSocketLogger) Warn(msg string, fields map[string]interface{}) {
	l.sendLogMessage("warn", msg, fields)
}

func (l *WebSocketLogger) Error(msg string, fields map[string]interface{}) {
	l.sendLogMessage("error", msg, fields)
}

// sendLogMessage sends a log message to all clients
func (l *WebSocketLogger) sendLogMessage(level, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// First print to console for server-side debugging
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if v, ok := fields["nodeID"].(string); ok {
		l.nodeID = v
	}
	if v, ok := fields["nodeId"].(string); ok {
		l.nodeID = v
	}
	fmt.Printf("[%s] %s: %s %v\n", level, l.nodeID, msg, fields)

	if l.persistentLogger != nil {
		l.persistentLogger.Write([]byte(msg))
		l.persistentLogger.Write([]byte{'\n'})
	}

	// Then broadcast via WebSocket
	l.wsManager.BroadcastMessage(MsgTypeLog, map[string]interface{}{
		"level":     level,
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"nodeId":    l.nodeID,
		"message":   msg,
		"fields":    fields,
	})
}

func (l *WebSocketLogger) Close() error {
	return l.persistentLogger.Close()
}

// ExecutionEventListener implements the engine.ExecutionListener interface
type ExecutionEventListener struct {
	wsManager *WebSocketManager
}

// NewExecutionEventListener creates a new execution event listener
func NewExecutionEventListener(wsManager *WebSocketManager) *ExecutionEventListener {
	return &ExecutionEventListener{
		wsManager: wsManager,
	}
}

// OnExecutionEvent handles execution events from the engine
func (l *ExecutionEventListener) OnExecutionEvent(event engine.ExecutionEvent) {
	// Map event types to WebSocket message types
	var msgType string

	switch event.Type {
	case engine.EventNodeStarted:
		msgType = MsgTypeNodeStart
	case engine.EventNodeCompleted:
		msgType = MsgTypeNodeComplete
	case engine.EventNodeError:
		msgType = MsgTypeNodeError
	case engine.EventValueProduced:
		msgType = MsgTypeDataFlow
	case engine.EventExecutionStart:
		msgType = MsgTypeExecStart
	case engine.EventExecutionEnd:
		msgType = MsgTypeExecEnd
	case engine.EventDebugData:
		msgType = MsgTypeDebugData
	default:
		msgType = MsgTypeExecStatus
	}

	// Create payload with timestamp
	payload := make(map[string]interface{})
	for k, v := range event.Data {
		payload[k] = v
	}
	payload["timestamp"] = event.Timestamp.Format(time.RFC3339Nano)
	if event.NodeID != "" {
		payload["nodeId"] = event.NodeID
	}

	// Broadcast the event
	l.wsManager.BroadcastMessage(msgType, payload)
}
