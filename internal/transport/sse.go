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
	connChan chan mcpsdk.Connection
	sessions map[string]*SSESession
	mu       sync.Mutex
}

// NewSSETransport creates a new SSE transport
func NewSSETransport() *SSETransport {
	return &SSETransport{
		connChan: make(chan mcpsdk.Connection),
		sessions: make(map[string]*SSESession),
	}
}

// Connect implements mcp.Transport
func (t *SSETransport) Connect(ctx context.Context) (mcpsdk.Connection, error) {
	select {
	case conn := <-t.connChan:
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
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

	session := &SSESession{
		id:       sessionID,
		msgChan:  make(chan mcpsdk.JSONRPCMessage, 10),
		readChan: make(chan mcpsdk.JSONRPCMessage, 10),
		done:     make(chan struct{}),
		cancel:   cancel,
	}

	// Register session
	t.mu.Lock()
	t.sessions[sessionID] = session
	t.mu.Unlock()

	log.Printf("SSE connection established: %s", sessionID)

	// Send endpoint event
	endpoint := fmt.Sprintf("/message?session=%s", sessionID)
	if err := t.sendEvent(w, flusher, "endpoint", endpoint); err != nil {
		log.Printf("Failed to send endpoint event: %v", err)
		return
	}

	// Notify Connect that a new connection is available
	// We do this in a goroutine so we don't block the SSE handler
	// But we need to ensure Connect picks it up.
	go func() {
		select {
		case t.connChan <- session:
			log.Printf("Session %s connected to server", sessionID)
		case <-ctx.Done():
			log.Printf("Session %s context done before connecting", sessionID)
		case <-time.After(5 * time.Second):
			log.Printf("Timeout waiting for server to accept session %s", sessionID)
		}
	}()

	// Keep connection alive and send messages
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer t.removeSession(sessionID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE connection closed: %s", sessionID)
			return

		case <-session.done:
			log.Printf("Session done: %s", sessionID)
			return

		case <-ticker.C:
			// Send keepalive
			if err := t.sendEvent(w, flusher, "ping", ""); err != nil {
				log.Printf("Failed to send keepalive: %v", err)
				return
			}

		case msg := <-session.msgChan:
			// Send message to client
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			if err := t.sendEvent(w, flusher, "message", string(data)); err != nil {
				log.Printf("Failed to send message: %v", err)
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
	t.mu.Lock()
	session, exists := t.sessions[sessionID]
	t.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Parse JSON-RPC request
	// We decode into JSONRPCRequest. If it's a notification, it should still work (ID will be empty/null).
	var request mcpsdk.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Send to session
	select {
	case session.readChan <- &request:
		// Accepted
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"accepted"}`))
	case <-session.done:
		http.Error(w, "Session closed", http.StatusServiceUnavailable)
	case <-time.After(5 * time.Second):
		http.Error(w, "Timeout processing request", http.StatusServiceUnavailable)
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
	t.mu.Lock()
	defer t.mu.Unlock()

	if session, exists := t.sessions[sessionID]; exists {
		session.Close()
		delete(t.sessions, sessionID)
		log.Printf("Session removed: %s", sessionID)
	}
}

// Close closes all sessions
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, session := range t.sessions {
		session.Close()
	}

	t.sessions = make(map[string]*SSESession)
	return nil
}

// SSESession implements mcpsdk.Connection
type SSESession struct {
	id       string
	msgChan  chan mcpsdk.JSONRPCMessage
	readChan chan mcpsdk.JSONRPCMessage
	done     chan struct{}
	cancel   context.CancelFunc
	once     sync.Once
}

func (s *SSESession) Read(ctx context.Context) (mcpsdk.JSONRPCMessage, error) {
	select {
	case msg := <-s.readChan:
		return msg, nil
	case <-s.done:
		return nil, fmt.Errorf("session closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SSESession) Write(ctx context.Context, msg mcpsdk.JSONRPCMessage) error {
	select {
	case s.msgChan <- msg:
		return nil
	case <-s.done:
		return fmt.Errorf("session closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *SSESession) Close() error {
	s.once.Do(func() {
		close(s.done)
		s.cancel()
	})
	return nil
}

func (s *SSESession) SessionID() string {
	return s.id
}
