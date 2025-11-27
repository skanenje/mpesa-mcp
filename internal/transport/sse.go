package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"
	"unsafe"

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

// makeJSONRPCID creates a jsonrpc2.ID using unsafe to set the unexported value field.
// This is necessary because the SDK's Int64ID and StringID functions are in an internal package.
func makeJSONRPCID(value any) mcpsdk.JSONRPCID {
	var id mcpsdk.JSONRPCID
	// Use reflection to set the unexported 'value' field
	v := reflect.ValueOf(&id).Elem()
	field := v.Field(0) // The 'value' field is the first (and only) field
	// Use unsafe to modify the unexported field
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
	return id
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
	log.Printf("[SSE] Received message request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Printf("[SSE] Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session ID
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		log.Printf("[SSE] Missing session parameter")
		http.Error(w, "Missing session parameter", http.StatusBadRequest)
		return
	}
	log.Printf("[SSE] Looking for session: %s", sessionID)

	// Find session
	t.mu.Lock()
	session, exists := t.sessions[sessionID]
	t.mu.Unlock()

	if !exists {
		log.Printf("[SSE] Session not found: %s (active sessions: %d)", sessionID, len(t.sessions))
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	log.Printf("[SSE] Session found: %s", sessionID)

	// Parse JSON-RPC request
	// Read the body first
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[SSE] Failed to read request body: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read request: %v", err), http.StatusBadRequest)
		return
	}

	// Unmarshal into a temporary structure to extract the raw ID value
	var rawMsg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(body, &rawMsg); err != nil {
		log.Printf("[SSE] Failed to unmarshal JSON: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate JSON-RPC version
	if rawMsg.JSONRPC != "2.0" {
		log.Printf("[SSE] Invalid JSON-RPC version: %s", rawMsg.JSONRPC)
		http.Error(w, "Invalid JSON-RPC version", http.StatusBadRequest)
		return
	}

	// Create the proper ID
	var id mcpsdk.JSONRPCID
	switch v := rawMsg.ID.(type) {
	case nil:
		// Notification (no ID)
		id = mcpsdk.JSONRPCID{}
	case float64:
		// Numeric ID - use unsafe to set the unexported value field
		id = makeJSONRPCID(int64(v))
	case string:
		// String ID - use unsafe to set the unexported value field
		id = makeJSONRPCID(v)
	default:
		log.Printf("[SSE] Invalid ID type: %T", rawMsg.ID)
		http.Error(w, "Invalid ID type", http.StatusBadRequest)
		return
	}

	// Construct the request
	request := &mcpsdk.JSONRPCRequest{
		ID:     id,
		Method: rawMsg.Method,
		Params: rawMsg.Params,
	}
	log.Printf("[SSE] Parsed JSON-RPC request - Method: %s, ID: %v", request.Method, request.ID)

	// Send to session
	select {
	case session.readChan <- request:
		// Accepted
		log.Printf("[SSE] Request accepted and queued for session %s", sessionID)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"accepted"}`))
	case <-session.done:
		log.Printf("[SSE] Session closed while processing request: %s", sessionID)
		http.Error(w, "Session closed", http.StatusServiceUnavailable)
	case <-time.After(5 * time.Second):
		log.Printf("[SSE] Timeout processing request for session: %s", sessionID)
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
	log.Printf("[Session %s] Waiting to read message...", s.id)
	select {
	case msg := <-s.readChan:
		log.Printf("[Session %s] Read message from client", s.id)
		return msg, nil
	case <-s.done:
		log.Printf("[Session %s] Read failed: session closed", s.id)
		return nil, fmt.Errorf("session closed")
	case <-ctx.Done():
		log.Printf("[Session %s] Read failed: context done", s.id)
		return nil, ctx.Err()
	}
}

func (s *SSESession) Write(ctx context.Context, msg mcpsdk.JSONRPCMessage) error {
	log.Printf("[Session %s] Writing message to client...", s.id)
	select {
	case s.msgChan <- msg:
		log.Printf("[Session %s] Message written successfully", s.id)
		return nil
	case <-s.done:
		log.Printf("[Session %s] Write failed: session closed", s.id)
		return fmt.Errorf("session closed")
	case <-ctx.Done():
		log.Printf("[Session %s] Write failed: context done", s.id)
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
