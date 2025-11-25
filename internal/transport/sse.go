package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SSETransport implements MCP transport over Server-Sent Events
type SSETransport struct {
	server      *mcpsdk.Server
	sessions    map[string]*Session
	sessionsMux sync.RWMutex
}

// Session represents an SSE connection session
type Session struct {
	ID       string
	writer   http.ResponseWriter
	flusher  http.Flusher
	ctx      context.Context
	cancel   context.CancelFunc
	messages chan *mcpsdk.JSONRPCMessage
	done     chan struct{}
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(server *mcpsdk.Server) *SSETransport {
	return &SSETransport{
		server:   server,
		sessions: make(map[string]*Session),
	}
}

// HandleSSE handles SSE connection requests
func (t *SSETransport) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Check if streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	// Create session
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(r.Context())

	session := &Session{
		ID:       sessionID,
		writer:   w,
		flusher:  flusher,
		ctx:      ctx,
		cancel:   cancel,
		messages: make(chan *mcpsdk.JSONRPCMessage, 10),
		done:     make(chan struct{}),
	}

	// Register session
	t.sessionsMux.Lock()
	t.sessions[sessionID] = session
	t.sessionsMux.Unlock()

	log.Printf("SSE connection established: %s", sessionID)

	// Send endpoint event
	endpoint := fmt.Sprintf("/message?session=%s", sessionID)
	t.sendEvent(w, flusher, "endpoint", endpoint)

	// Keep connection alive and send messages
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE connection closed: %s", sessionID)
			t.removeSession(sessionID)
			return

		case <-ticker.C:
			// Send keepalive
			if err := t.sendEvent(w, flusher, "ping", ""); err != nil {
				log.Printf("Failed to send keepalive: %v", err)
				t.removeSession(sessionID)
				return
			}

		case msg := <-session.messages:
			// Send message to client
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			if err := t.sendEvent(w, flusher, "message", string(data)); err != nil {
				log.Printf("Failed to send message: %v", err)
				t.removeSession(sessionID)
				return
			}
		}
	}
}

// HandleMessage handles incoming JSON-RPC messages
func (t *SSETransport) HandleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session ID
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Missing session parameter", http.StatusBadRequest)
		return
	}

	// Find session
	t.sessionsMux.RLock()
	session, exists := t.sessions[sessionID]
	t.sessionsMux.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Parse JSON-RPC request
	var request mcpsdk.JSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Received request: method=%s, id=%v", request.Method, request.ID)

	// Process request asynchronously
	go t.processRequest(session, &request)

	// Return 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

// processRequest processes a JSON-RPC request and sends response via SSE
func (t *SSETransport) processRequest(session *Session, request *mcpsdk.JSONRPCMessage) {
	ctx := session.ctx

	// Handle the request using the MCP server
	response := t.server.HandleMessage(ctx, request)

	// Send response back via SSE
	select {
	case session.messages <- response:
		log.Printf("Response sent for request id=%v", request.ID)
	case <-ctx.Done():
		log.Printf("Session closed before response could be sent")
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending response for request id=%v", request.ID)
	}
}

// sendEvent sends an SSE event
func (t *SSETransport) sendEvent(w http.ResponseWriter, flusher http.Flusher, event, data string) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}

	if data != "" {
		if _, err := fmt.Fprintf(w, "data: %s\n", data); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\n"); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// removeSession removes a session from the transport
func (t *SSETransport) removeSession(sessionID string) {
	t.sessionsMux.Lock()
	defer t.sessionsMux.Unlock()

	if session, exists := t.sessions[sessionID]; exists {
		session.cancel()
		close(session.messages)
		delete(t.sessions, sessionID)
		log.Printf("Session removed: %s", sessionID)
	}
}

// Close closes all sessions
func (t *SSETransport) Close() {
	t.sessionsMux.Lock()
	defer t.sessionsMux.Unlock()

	for _, session := range t.sessions {
		session.cancel()
		close(session.messages)
	}

	t.sessions = make(map[string]*Session)
	log.Println("All sessions closed")
}
