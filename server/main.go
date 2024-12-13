package main

import (
    "encoding/json"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "golang.org/x/exp/slog"
)

// Message types for WebSocket communication
type MessageType string

const (
    TypeInitialState MessageType = "initial_state"
    TypePaddleUpdate MessageType = "paddle_update"
)

// Message structure for WebSocket communication
type Message struct {
    Type    MessageType  `json:"type"`
    Payload interface{} `json:"payload"`
}

// PaddlePosition represents the position of a paddle
type PaddlePosition struct {
    Side string  `json:"side"`    // "left" or "right"
    Y    float64 `json:"y"`       // Y coordinate
}

// GameState represents the current state of the game
type GameState struct {
    LeftPaddle  float64 `json:"leftPaddle"`
    RightPaddle float64 `json:"rightPaddle"`
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all connections for now 🦍
    },
}

type Server struct {
    // Mutex to protect connections and game state
    sync.RWMutex
    // Connections store
    connections map[*websocket.Conn]bool
    // Add connection count for metrics
    connectionCount int
    // Game state
    gameState GameState
}

func NewServer() *Server {
    return &Server{
        connections: make(map[*websocket.Conn]bool),
        gameState: GameState{
            LeftPaddle:  300, // Initial positions
            RightPaddle: 300,
        },
    }
}

// Broadcast sends a message to all connected clients
func (s *Server) broadcast(msg Message) {
    s.RLock()
    for conn := range s.connections {
        if err := conn.WriteJSON(msg); err != nil {
            slog.Error("Failed to send message",
                "error", err,
                "addr", conn.RemoteAddr(),
                "timestamp", time.Now().Format(time.RFC3339))
        }
    }
    s.RUnlock()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
    // Log incoming connection attempt
    slog.Info("Incoming WebSocket connection attempt",
        "remote_addr", r.RemoteAddr,
        "user_agent", r.UserAgent(),
        "timestamp", time.Now().Format(time.RFC3339))

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("Failed to upgrade connection",
            "error", err,
            "remote_addr", r.RemoteAddr,
            "timestamp", time.Now().Format(time.RFC3339))
        return
    }

    // Add connection to our map
    s.Lock()
    s.connections[conn] = true
    s.connectionCount++
    currentCount := s.connectionCount
    s.Unlock()

    slog.Info("New connection established",
        "addr", conn.RemoteAddr(),
        "total_connections", currentCount,
        "timestamp", time.Now().Format(time.RFC3339))

    // Send initial game state
    initialMsg := Message{
        Type:    TypeInitialState,
        Payload: s.gameState,
    }
    if err := conn.WriteJSON(initialMsg); err != nil {
        slog.Error("Failed to send initial state",
            "error", err,
            "addr", conn.RemoteAddr(),
            "timestamp", time.Now().Format(time.RFC3339))
    }

    // Remove connection when function returns
    defer func() {
        s.Lock()
        delete(s.connections, conn)
        s.connectionCount--
        currentCount := s.connectionCount
        s.Unlock()
        conn.Close()
        slog.Info("Connection closed",
            "addr", conn.RemoteAddr(),
            "remaining_connections", currentCount,
            "timestamp", time.Now().Format(time.RFC3339))
    }()

    // Handle incoming messages
    for {
        var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                slog.Error("WebSocket error",
                    "error", err,
                    "addr", conn.RemoteAddr(),
                    "timestamp", time.Now().Format(time.RFC3339))
            }
            break
        }

        // Handle paddle updates
        if msg.Type == TypePaddleUpdate {
            if pos, ok := msg.Payload.(map[string]interface{}); ok {
                s.Lock()
                if side, ok := pos["side"].(string); ok {
                    if y, ok := pos["y"].(float64); ok {
                        if side == "left" {
                            s.gameState.LeftPaddle = y
                        } else {
                            s.gameState.RightPaddle = y
                        }
                        // Broadcast the update to all clients
                        s.broadcast(msg)
                    }
                }
                s.Unlock()
            }
        }
    }
}

func main() {
    // Setup JSON logger with timestamp
    logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true,
    })
    logger := slog.New(logHandler)
    slog.SetDefault(logger)

    server := NewServer()

    // Log server configuration
    slog.Info("🦍 STRONK SERVER CONFIGURATION 🦍",
        "port", 42069,
        "timestamp", time.Now().Format(time.RFC3339),
        "version", "1.0.0",
        "log_level", "debug")

    // Serve static files from /app/src directory
    fs := http.FileServer(http.Dir("/app/src"))
    http.Handle("/", http.StripPrefix("/", fs))

    // Handle WebSocket connections
    http.HandleFunc("/ws", server.handleWS)

    slog.Info("🦍 STRONK SERVER STARTING ON PORT 42069 🦍",
        "timestamp", time.Now().Format(time.RFC3339))
    if err := http.ListenAndServe(":42069", nil); err != nil {
        slog.Error("Server failed to start",
            "error", err,
            "timestamp", time.Now().Format(time.RFC3339))
        os.Exit(1)
    }
}
