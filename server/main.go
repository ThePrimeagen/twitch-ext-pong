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
    TypeTeamAssign   MessageType = "team_assign"    // New type for team assignment
)

// Message structure for WebSocket communication
type Message struct {
    Type    MessageType     `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// TeamAssignment represents team assignment for a player
type TeamAssignment struct {
    Team string `json:"team"`  // "left" or "right"
}

// PaddlePosition represents the position of a paddle
type PaddlePosition struct {
    Side string  `json:"side"`    // "left" or "right"
    Y    float64 `json:"y"`       // Y coordinate
}

// Validate ensures paddle position is within bounds
func (p *PaddlePosition) Validate() error {
    if p.Y < 0 || p.Y > 600 { // Canvas height validation
        return fmt.Errorf("invalid paddle Y position: %f", p.Y)
    }
    return nil
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
    defer s.RUnlock()

    deadConns := make([]*websocket.Conn, 0)
    for conn := range s.connections {
        if err := conn.WriteJSON(msg); err != nil {
            slog.Error("Failed to broadcast message",
                "error", err,
                "addr", conn.RemoteAddr(),
                "timestamp", time.Now().Format(time.RFC3339))
            deadConns = append(deadConns, conn)
        }
    }

    // Clean up dead connections outside the read lock
    if len(deadConns) > 0 {
        s.Lock()
        for _, conn := range deadConns {
            delete(s.connections, conn)
            s.connectionCount--
            slog.Info("🦍 REMOVED DEAD CONNECTION 🦍",
                "addr", conn.RemoteAddr(),
                "remaining", s.connectionCount,
                "timestamp", time.Now().Format(time.RFC3339))
        }
        s.Unlock()
    }
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
            "addr", conn.RemoteAddr,
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
            var paddlePos PaddlePosition
            if err := json.Unmarshal(msg.Payload, &paddlePos); err != nil {
                slog.Error("Failed to parse paddle position",
                    "error", err,
                    "addr", conn.RemoteAddr(),
                    "timestamp", time.Now().Format(time.RFC3339))
                continue
            }

            // Validate paddle position
            if err := paddlePos.Validate(); err != nil {
                slog.Error("Invalid paddle position",
                    "error", err,
                    "addr", conn.RemoteAddr(),
                    "timestamp", time.Now().Format(time.RFC3339))
                continue
            }

            s.Lock()
            // All players control left paddle
            s.gameState.LeftPaddle = paddlePos.Y
            s.Unlock()

            // Broadcast outside of lock
            s.broadcast(msg)

            slog.Info("🦍 PADDLE MOVED 🦍",
                "side", paddlePos.Side,
                "y", paddlePos.Y,
                "timestamp", time.Now().Format(time.RFC3339))
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
