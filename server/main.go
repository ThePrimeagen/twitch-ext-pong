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

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all connections for now 🦍
    },
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
    // Message types
    TypeInitialState MessageType = "initial_state"
    TypePaddleUpdate MessageType = "paddle_update"
    TypePlayerAssign MessageType = "player_assign"
)

// Message represents a WebSocket message with type and payload
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
    LeftPaddle  float64 `json:"left_paddle"`
    RightPaddle float64 `json:"right_paddle"`
}

// PlayerAssignment represents which side a player is assigned to
type PlayerAssignment struct {
    Side string `json:"side"` // "left" or "right"
}

type Server struct {
    sync.RWMutex
    connections     map[*websocket.Conn]bool
    connectionCount int
    gameState      GameState
    playerSides    map[*websocket.Conn]string
}

func NewServer() *Server {
    return &Server{
        connections: make(map[*websocket.Conn]bool),
        gameState: GameState{
            LeftPaddle:  0,
            RightPaddle: 0,
        },
        playerSides: make(map[*websocket.Conn]string),
    }
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
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

    s.Lock()
    s.connections[conn] = true
    s.connectionCount++
    currentCount := s.connectionCount
    s.Unlock()

    slog.Info("New connection established",
        "addr", conn.RemoteAddr(),
        "total_connections", currentCount,
        "timestamp", time.Now().Format(time.RFC3339))

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

    // Send initial game state
    initialState := Message{
        Type: TypeInitialState,
        Payload: s.gameState,
    }
    if err := conn.WriteJSON(initialState); err != nil {
        slog.Error("Failed to send initial state",
            "error", err,
            "addr", conn.RemoteAddr(),
            "timestamp", time.Now().Format(time.RFC3339))
    }

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            slog.Debug("Connection read error",
                "error", err,
                "addr", conn.RemoteAddr(),
                "timestamp", time.Now().Format(time.RFC3339))
            break
        }

        var paddlePos PaddlePosition
        if err := json.Unmarshal(message, &paddlePos); err != nil {
            slog.Error("Failed to parse paddle position",
                "error", err,
                "data", string(message))
            continue
        }

        s.Lock()
        if paddlePos.Side == "left" {
            s.gameState.LeftPaddle = paddlePos.Y
        } else if paddlePos.Side == "right" {
            s.gameState.RightPaddle = paddlePos.Y
        }
        s.Unlock()

        update := Message{
            Type:    TypePaddleUpdate,
            Payload: paddlePos,
        }
        s.broadcast(update)
    }
}

func (s *Server) broadcast(message Message) {
    s.RLock()
    defer s.RUnlock()

    for conn := range s.connections {
        if err := conn.WriteJSON(message); err != nil {
            slog.Error("Failed to broadcast",
                "error", err,
                "addr", conn.RemoteAddr())
        }
    }
}

func main() {
    logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true,
    })
    logger := slog.New(logHandler)
    slog.SetDefault(logger)

    server := NewServer()

    slog.Info("🦍 STRONK SERVER CONFIGURATION 🦍",
        "port", 42069,
        "timestamp", time.Now().Format(time.RFC3339),
        "version", "1.0.0",
        "log_level", "debug")

    fs := http.FileServer(http.Dir("/app/src"))
    http.Handle("/", http.StripPrefix("/", fs))

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
